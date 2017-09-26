package http

import (
	"errors"
	gohttp "net/http"
	"os"
)

func LocalWWWFileSystem(root string) (gohttp.FileSystem, error) {

	info, err := os.Stat(root)

	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("root is not a directory")
	}

	fs := gohttp.Dir(root)
	return fs, nil
}

func LocalWWWHandler(fs gohttp.FileSystem) (gohttp.Handler, error) {
	return gohttp.FileServer(fs), nil
}
