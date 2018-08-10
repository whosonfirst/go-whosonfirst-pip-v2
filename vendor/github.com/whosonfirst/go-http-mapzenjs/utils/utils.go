package utils

import (
	"github.com/whosonfirst/go-http-mapzenjs"
	"net/http"
)

// TO DO - options to allow toggling map styles

func AppendMapzenJSAssets(mux *http.ServeMux) error {

	handler, err := mapzenjs.MapzenJSAssetsHandler()

	if err != nil {
		return err
	}

	return AssignMapzenJSAssetsURLs(mux, handler)
}

func AssignMapzenJSAssetsURLs(mux *http.ServeMux, handler http.Handler) error {

	mux.Handle("/javascript/mapzen.min.js", handler)
	mux.Handle("/javascript/mapzen.js", handler)
	mux.Handle("/javascript/tangram.min.js", handler)
	mux.Handle("/javascript/tangram.js", handler)
	mux.Handle("/css/mapzen.js.css", handler)
	mux.Handle("/tangram/refill-style.zip", handler)
	mux.Handle("/tangram/refill-style-themes-label.zip", handler)

	return nil
}
