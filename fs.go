package images

import (
	"image"
	"io"
	"os"
	"path"
	"strings"
)

type FileSystem interface {
	Create(string) (io.WriteCloser, error)
	Open(string) (io.ReadCloser, error)
}

type LocalFileSystem struct{}

func (fs LocalFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (fs LocalFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

var fs FileSystem = LocalFileSystem{}

func Open(filename string) (image.Image, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Decode(file)
}

func Close(img image.Image, filename string, quality int) error {
	file, err := fs.Create(filename)
	if err != nil {
		return err
	}
	ext := path.Ext(filename)
	err = Encode(file, img, strings.ReplaceAll(ext, ".", ""), quality)
	if err != nil {
		return err
	}
	err = file.Close()
	return err
}
