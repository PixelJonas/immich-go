package local

import (
	"context"
	"immich-go/fshelper"
	"immich-go/host/host_if"

	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

type Connection struct {
	ctx  context.Context
	path string
}

func New(ctx context.Context, u *url.URL) (host_if.HostConnection, error) {
	return &Connection{
		path: u.Path,
		ctx:  ctx,
	}, nil
}

func (c Connection) Close() error {
	return nil
}

func (Connection) CommandContext(ctx context.Context, name string, args ...string) (host_if.Runner, error) {
	return exec.CommandContext(ctx, name, args...), nil
}

type localRwFS struct {
	path string
	fs.FS
}

func (c Connection) OpenFS() (fshelper.RemoteFS, error) {
	return FS(c.path)
}

func FS(path string) (*localRwFS, error) {
	fsys := localRwFS{
		path: path,
		FS:   os.DirFS(path),
	}
	return &fsys, nil
}

func (fsys localRwFS) Remove(name string) error {
	return os.Remove(filepath.Join(fsys.path, name))
}
func (fsys localRwFS) Create(name string) (fshelper.Writer, error) {
	return os.Create(filepath.Join(fsys.path, name))
}
func (fsys localRwFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(fsys.path, name))
}

func (fsys localRwFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(fsys.path, name))
}

func (fsys localRwFS) Close() error { return nil }
