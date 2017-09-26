package http

import (
	gohttp "net/http"
)

func WWWFileSystem() gohttp.FileSystem {
	return assetFS()
}

func WWWHandler() (gohttp.Handler, error) {

	fs := assetFS()
	return gohttp.FileServer(fs), nil
}
