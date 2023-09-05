package fshelper

import (
	"errors"
	"io/fs"
)

type Remover interface {
	Remove(name string) error
}

type ReadDirer interface {
	ReadDir(name string) ([]fs.DirEntry, error)
}

type Creator interface {
	Create(name string) (Writer, error)
}
type RemoteFS interface {
	fs.FS
	Remover
	ReadDirer
	Creator
	Close() error
}

func Create(fsys fs.FS, name string) (Writer, error) {
	if fsys, ok := fsys.(Creator); ok {
		return fsys.Create(name)
	}
	return nil, errors.New("Create method not implemented")
}

func ReadDir(fsys fs.FS, name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(fsys, name)
}

func Remove(fsys fs.FS, name string) error {
	if fsys, ok := fsys.(Remover); ok {
		return fsys.Remove(name)
	}
	return nil
}
