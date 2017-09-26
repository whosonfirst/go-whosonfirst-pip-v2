package http

import (
	gohttp "net/http"
)

func BundledWWWFileSystem() (gohttp.FileSystem, error) {
	fs := assetFS()
	return fs, nil
}

func BundledWWWHandler() (gohttp.Handler, error) {

	fs := assetFS()
	return gohttp.FileServer(fs), nil
}
