package mapzenjs

import (
	"net/http"
)

func MapzenJSHandler() (http.Handler, error) {

	fs := assetFS()
	return http.FileServer(fs), nil
}
