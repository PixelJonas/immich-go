package cmdmetadata

import (
	"context"
	"flag"
	"immich-go/host"
	"immich-go/host/host_if"
	"immich-go/immich"
	"immich-go/immich/logger"
	"immich-go/immich/metadata"
	"math"
	"path"
	"strings"
	"time"
)

type MetadataCmd struct {
	Immich         *immich.ImmichClient // Immich client
	Log            *logger.Logger
	DryRun         bool
	UseNameAsDate  bool
	MissingDate    bool
	Host           string
	HostConnection host_if.HostConnection
}

func NewMetadataCmd(ctx context.Context, ic *immich.ImmichClient, logger *logger.Logger, args []string) (*MetadataCmd, error) {
	var err error
	cmd := flag.NewFlagSet("metadata", flag.ExitOnError)
	app := MetadataCmd{
		Immich: ic,
		Log:    logger,
	}

	cmd.StringVar(&app.Host, "host-access", "", "Immich's direct access to the asset storage ssh://user:password@host:port, scp://user:password@host:port/path/to/the/library, smb://user:password@host:port/path/to/the/library")
	cmd.BoolVar(&app.DryRun, "dry-run", true, "display actions, but don't touch the server assets")
	cmd.BoolVar(&app.MissingDate, "missing-date", false, "select all assets where the date is missing")
	cmd.BoolVar(&app.UseNameAsDate, "use-name-as-date", false, "select check assets against their filename, and update the immich date when needed")
	err = cmd.Parse(args)
	return &app, err
}

func MetadataCommand(ctx context.Context, ic *immich.ImmichClient, log *logger.Logger, args []string) error {
	app, err := NewMetadataCmd(ctx, ic, log, args)
	if err != nil {
		return err
	}

	app.HostConnection, err = host.Open(ctx, app.Host)

	var dockerConn *docker.DockerConnect

	dockerConn, err = docker.NewDockerConnection(ctx, app.HostConnection, "immich_server")
	if err != nil {
		return err
	}

	app.Log.OK("Connected to the immich's docker container at %q", app.HostConnection)

	app.Log.MessageContinue(logger.OK, "Get server's assets...")
	list, err := app.Immich.GetAllAssets(ctx, nil)
	if err != nil {
		return err
	}
	app.Log.MessageTerminate(logger.OK, " %d received", len(list))

	type broken struct {
		a *immich.Asset
		metadata.SideCar
		fixable bool
		reason  []string
	}

	now := time.Now().Add(time.Hour * 24)
	brockenAssets := []broken{}
	for _, a := range list {
		ba := broken{a: a}

		if (app.MissingDate) && a.ExifInfo.DateTimeOriginal.IsZero() {
			ba.reason = append(ba.reason, "capture date not set")
		}
		if (app.MissingDate) && (a.ExifInfo.DateTimeOriginal.Year() < 1900 || a.ExifInfo.DateTimeOriginal.Compare(now) > 0) {
			ba.reason = append(ba.reason, "capture date invalid")
		}

		if app.UseNameAsDate {
			dt := metadata.TakeTimeFromName(path.Base(a.OriginalPath))
			if !dt.IsZero() {
				if a.ExifInfo.DateTimeOriginal.IsZero() || (math.Abs(float64(dt.Sub(a.ExifInfo.DateTimeOriginal))) > float64(24.0*time.Hour)) {
					ba.reason = append(ba.reason, "capture date invalid, but the name contains a date")
					ba.fixable = true
					ba.SideCar.DateTaken = dt
				}
			}
		}

		/*
			if a.ExifInfo.Latitude == nil || a.ExifInfo.Longitude == nil {
				ba.reason = append(ba.reason, "GPS coordinates not set")
			} else if math.Abs(*a.ExifInfo.Latitude) < 0.00001 && math.Abs(*a.ExifInfo.Longitude) < 0.00001 {
				ba.reason = append(ba.reason, "GPS coordinates is near of 0;0")
			}
		*/
		if len(ba.reason) > 0 {
			brockenAssets = append(brockenAssets, ba)
		}
	}

	fixable := 0
	for _, b := range brockenAssets {
		if b.fixable {
			fixable++
		}
		app.Log.OK("%s, (%s %s): %s", b.a.OriginalPath, b.a.ExifInfo.Make, b.a.ExifInfo.Model, strings.Join(b.reason, ", "))
	}
	app.Log.OK("%d broken assets", len(brockenAssets))
	app.Log.OK("Among them, %d can be fixed with current settings", fixable)

	if fixable == 0 {
		return nil
	}

	if app.DryRun {
		log.OK("Dry-run mode. Exiting")
		log.OK("use -dry-run=false after metadata command")
		return nil
	}

	uploader, err := dockerConn.BatchUpload(ctx, "/usr/src/app")
	if err != nil {
		return err
	}

	defer uploader.Close()

	for _, b := range brockenAssets {
		if !b.fixable {
			continue
		}
		a := b.a
		scContent, err := b.SideCar.Bytes()
		if err != nil {
			return err
		}
		err = uploader.Upload(a.OriginalPath+".xmp", scContent)
		if err != nil {
			return err
		}
		app.Log.OK("Uploaded sidecar for %s... ", a.OriginalPath)
	}

	err = app.SidecarSync(ctx)
	if err != nil {
		return err
	}

	err = app.SidecarDiscover(ctx)
	return err
}

func (app *MetadataCmd) SidecarSync(ctx context.Context) error {
	app.Log.MessageContinue(logger.OK, "Start SidecarSync job....")
	defer app.Log.MessageTerminate(logger.OK, "terminated")
	return app.Immich.StartAndWaitJob(ctx, immich.JobSidecar, immich.CmdStart, true)
}

func (app *MetadataCmd) SidecarDiscover(ctx context.Context) error {
	app.Log.MessageContinue(logger.OK, "Start SidecarDiscover job....")
	defer app.Log.MessageTerminate(logger.OK, "terminated")
	return app.Immich.StartAndWaitJob(ctx, immich.JobSidecar, immich.CmdStart, false)
}
