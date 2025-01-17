package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

type NameResolver interface {
	ResolveName(la *LocalAssetFile) (string, error)
}

type GooglePhotosAssetBrowser struct {
	fs.FS
	albums map[string]string
}

func BrowseGooglePhotosAssets(fsys fs.FS) *GooglePhotosAssetBrowser {
	return &GooglePhotosAssetBrowser{
		FS: fsys,
	}
}

// browseGooglePhotos collects and filters assets from a file systems (fs.FS) to create a channel of localFile.
// The function scans all given file systems and processes JSON metadata files to extract relevant assets.

func (fsys *GooglePhotosAssetBrowser) Browse(ctx context.Context) chan *LocalAssetFile {
	fileChan := make(chan *LocalAssetFile)

	// Start a goroutine to browse the FS and collect the list of files
	go func(ctx context.Context) {
		defer close(fileChan) // Close the channel when the goroutine finishes

		err := fs.WalkDir(fsys, ".",
			func(name string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Check if the context has been cancelled
				select {
				case <-ctx.Done():
					// If the context has been cancelled, return immediately
					return ctx.Err()
				default:
				}

				if d.IsDir() {
					return nil
				}

				if strings.ToLower(path.Ext(name)) != ".json" {
					return nil
				}

				md, err := readJSON[googleMetaData](fsys, name)
				if err == nil && md != nil && len(md.URL) > 0 {
					ext := strings.ToLower(path.Ext(md.Title))
					switch ext {
					case ".3fr",
						".ari",
						".arw",
						".avif",
						".cap",
						".cin",
						".cr2",
						".cr3",
						".crw",
						".dcr",
						".dng",
						".erf",
						".fff",
						".gif",
						".heic",
						".heif",
						".iiq",
						".insp",
						".jpeg",
						".jpg",
						".jxl",
						".k25",
						".kdc",
						".mrw",
						".nef",
						".orf",
						".ori",
						".pef",
						".png",
						".raf",
						".raw",
						".rwl",
						".sr2",
						".srf",
						".srw",
						".tif",
						".tiff",
						".webp",
						".x3f",
						".3gp",
						".avi",
						".flv",
						".insv",
						".m2ts",
						".mkv",
						".mov",
						".mp4",
						".mpg",
						".mts",
						".webm",
						".wmv":
					case "":
						// Few titles don't have extension. Assume .jpg
						ext = ".jpg"
						md.Title += ext
					default:
						return nil
					}

					dir := path.Dir(name)

					if path.Base(dir) == "Failed Videos" {
						return nil
					}

					f := LocalAssetFile{
						FSys:         fsys,
						FileName:     path.Join(dir, nameReplacer.Replace(md.Title)),
						Title:        md.Title,
						Trashed:      md.Trashed,
						FromPartner:  md.GooglePhotosOrigin.FromPartnerSharing != nil,
						DateTaken:    md.PhotoTakenTime.Time(),
						Latitude:     md.GeoData.Latitude,
						Longitude:    md.GeoData.Longitude,
						Altitude:     md.GeoData.Altitude,
						MetadataFile: name,
					}

					if album, ok := fsys.albums[dir]; ok {
						f.AddAlbum(album)
					}

					// Check if the context has been cancelled before sending the file
					select {
					case <-ctx.Done():
						// If the context has been cancelled, return immediately
						return ctx.Err()
					case fileChan <- &f:
					}
				}
				return nil // ignore json errors...
			})

		if err != nil {
			// Check if the context has been cancelled before sending the error
			select {
			case <-ctx.Done():
				// If the context has been cancelled, return immediately
				return
			case fileChan <- &LocalAssetFile{
				Err: err,
			}:
			}
		}
	}(ctx)

	return fileChan
}

func (fsys *GooglePhotosAssetBrowser) ResolveName(la *LocalAssetFile) (string, error) {
	if la.isNotResolvable {
		return "not resolved", fmt.Errorf("%s: %w", la.Title, os.ErrNotExist)
	}
	if la.isResolved {
		return la.FileName, nil
	}
	ext := path.Ext(la.Title)
	base := strings.TrimSuffix(la.Title, ext)
	dir := path.Dir(la.FileName)
	baseUnicode := utf16.Encode([]rune(base))

	if len(baseUnicode) > 46 {
		baseUnicode = baseUnicode[:46]
		base = string(utf16.Decode(baseUnicode))
	}
	pattern := nameReplacer.Replace(base) + "*.*"

	matches, err := fs.Glob(fsys, path.Join(dir, pattern))
	if err != nil {
		return "", fmt.Errorf("can't resolve name: %w", err)
	}

	ext = strings.ToLower(ext)

	for _, m := range matches {
		if strings.Compare(ext, strings.ToLower(path.Ext(m))) == 0 {
			la.FileName = m
			la.fInfo, err = fs.Stat(fsys, m)
			if err != nil {
				return "", err
			}
			la.isResolved = true
			return m, nil
		}
	}
	return "", fmt.Errorf("can't find the image with title %q, pattern: %q: %w", la.Title, pattern, os.ErrNotExist)
}

var nameReplacer = strings.NewReplacer(" ", "?", "/", "?", ":", "?", "&", "?")

type googleMetaData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// ImageViews         string         `json:"imageViews"`
	// CreationTime       googTimeObject `json:"creationTime"`
	PhotoTakenTime     googTimeObject `json:"photoTakenTime"`
	GeoData            googGeoData    `json:"geoData"`
	GeoDataExif        googGeoData    `json:"geoDataExif"`
	Trashed            bool           `json:"trashed,omitempty"`
	Archived           bool           `json:"archived,omitempty"`
	URL                string         `json:"url"`
	GooglePhotosOrigin struct {
		MobileUpload struct {
			DeviceFolder struct {
				LocalFolderName string `json:"localFolderName"`
			} `json:"deviceFolder"`
			DeviceType string `json:"deviceType"`
		} `json:"mobileUpload"`
		FromPartnerSharing *struct {
		} `json:"fromPartnerSharing"`
	} `json:"googlePhotosOrigin"`
}

type googGeoData struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Altitude      float64 `json:"altitude"`
	LatitudeSpan  float64 `json:"latitudeSpan"`
	LongitudeSpan float64 `json:"longitudeSpan"`
}

type googTimeObject struct {
	Timestamp int64 `json:"timestamp"`
	// Formatted string    `json:"formatted"`
}

func (gt googTimeObject) Time() time.Time {
	t := time.Unix(gt.Timestamp, 0)
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return t
}

func (t *googTimeObject) UnmarshalJSON(data []byte) error {
	type Alias googTimeObject
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	t.Timestamp, err = strconv.ParseInt(aux.Timestamp, 10, 64)

	return err
}

func (fsys *GooglePhotosAssetBrowser) BrowseAlbums(ctx context.Context) error {
	fsys.albums = map[string]string{}

	err := fs.WalkDir(fsys, ".",
		func(name string, d fs.DirEntry, err error) error {
			type MetaData struct {
				Title string `json:"title"`
			}

			if err != nil {
				return err
			}

			// Check if the context has been cancelled
			select {
			case <-ctx.Done():
				// If the context has been cancelled, return immediately
				return ctx.Err()
			default:
			}

			if d.IsDir() {
				return nil
			}

			base := path.Base(name)

			// Localized metadata file names according bard. https://g.co/bard/share/bcc70cb206e2
			switch base {
			case "metadata.json",
				"métadonnées.json",
				"Metadaten.json",
				"metadatos.json",
				"metadati.json",
				"metadados.json",
				"метаданные.json",
				"メタデータ.json",
				"元数据.json",
				"Metadata.json":
			default:
				return nil
			}

			md, err := readJSON[MetaData](fsys, name)
			if err != nil || md.Title == "" {
				return nil
			}
			fsys.albums[path.Dir(name)] = md.Title
			return nil
		})

	return err

}
