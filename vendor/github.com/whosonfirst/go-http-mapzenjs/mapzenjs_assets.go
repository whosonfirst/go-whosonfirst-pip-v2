package mapzenjs

import (
	"net/http"
)

func MapzenJSAssetsHandler() (http.Handler, error) {

	fs := assetFS()
	return http.FileServer(fs), nil
}
