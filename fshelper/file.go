package fshelper

import (
	"errors"
	"io/fs"
)

type Writer interface {
	fs.File
	Write([]byte) (int, error)
}

type WriteStringer interface {
	fs.File
	WriteString(string) (int, error)
}

func Write(f fs.File, b []byte) (int, error) {
	if f, ok := f.(Writer); ok {
		return f.Write(b)
	}
	return 0, errors.New("Write method not implemented")
}

func WriteString(f Writer, s string) (int, error) {

	if f, ok := f.(WriteStringer); ok {
		return f.WriteString(s)
	}
	return f.Write([]byte(s))
}
