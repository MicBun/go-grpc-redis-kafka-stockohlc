package filesystem

import (
	"io/fs"
	"os"
)

type OsFS struct{}

type FS interface {
	ReadDir(dirname string) ([]os.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

func NewOsFS() *OsFS {
	return &OsFS{}
}

func (OsFS) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (OsFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name) // #nosec
}

type DirEntry interface {
	Name() string
	IsDir() bool
	Type() fs.FileMode
	Info() (fs.FileInfo, error)
}
