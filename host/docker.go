package host

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"immich-go/host/host_if"
	"strings"
)

type Docker struct {
	Host         host_if.HostConnection
	UploadDir    string
	ImmichServer string
}

func NewDockerConnection(ctx context.Context, host host_if.HostConnection) (*Docker, error) {
	dc := Docker{
		Host: host,
	}

	l, err := dc.DockerPS(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range l {
		if c == "immich_server" {
			dc.ImmichServer = c
		}
	}

	return &dc, nil
}

func (dc Docker) DockerPS(ctx context.Context) ([]string, error) {
	cmd, err := dc.Host.CommandContext(ctx, "docker", "ps", "--format", "{{.Names}}")
	if err != nil {
		return nil, err
	}
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	r := []string{}
	buff := bytes.NewBuffer(b)
	for {
		l, err := buff.ReadString('\n')
		l = strings.TrimSuffix(l, "\n")
		if err != nil {
			break
		}
		r = append(r, l)
	}
	return r, nil
}

type mounts struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	Rw          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

func (dc Docker) ImmichLibrary(ctx context.Context) (string, error) {
	cmd, err := dc.Host.CommandContext(ctx, "docker", "inspect", "--format", "'{{json .Mounts}}'", dc.ImmichServer)
	if err != nil {
		return "", err
	}
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Join(err, errors.New(string(b)))
	}
	mounts, err := decodeJson[[]mounts](b)
	if err != nil {
		return "", err
	}

	for _, m := range mounts {
		if m.Destination == "/usr/src/app/upload" {
			dc.UploadDir = m.Source
		}
	}
	if dc.UploadDir == "" {
		return "", errors.New("Can't detect immich library on the host")
	}
	return dc.UploadDir, nil
}

func decodeJson[T any](b []byte) (T, error) {
	var r T
	err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&r)
	return r, err
}

// func (c DockerConnection) RunJson(cmd Commander, resp any) error {
// 	b, err := cmd.Output()
// 	if err != nil {
// 		return err
// 	}
// 	return json.NewDecoder(bytes.NewBuffer(b)).Decode(resp)
// }
