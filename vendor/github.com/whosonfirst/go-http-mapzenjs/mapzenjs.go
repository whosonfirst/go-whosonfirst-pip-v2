package mapzenjs

import (
	"bufio"
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	_ "log"
	"net/http"
	"net/http/httptest"
	"strconv"
)

type MapzenJSOptions struct {
	AppendAPIKey bool
	AppendJS     bool
	AppendCSS    bool
	APIKey       string
	JS           []string
	CSS          []string
}

func DefaultMapzenJSOptions() *MapzenJSOptions {

	opts := MapzenJSOptions{
		AppendAPIKey: true,
		AppendJS:     true,
		AppendCSS:    true,
		APIKey:       "mapzen-xxxxxx",
		JS:           []string{"/javascript/mapzen.min.js"},
		CSS:          []string{"/css/mapzen.js.css"},
	}

	return &opts
}

func MapzenJSHandler(handler http.Handler, opts *MapzenJSOptions) (http.Handler, error) {

	h := MapzenJSWriter{
		handler: handler,
		options: opts,
	}

	return h, nil
}

type MapzenJSWriter struct {
	handler http.Handler
	options *MapzenJSOptions
}

func (h MapzenJSWriter) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {

	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)

	body := rec.Body.Bytes()
	reader := bytes.NewReader(body)
	doc, err := html.Parse(reader)

	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	var f func(node *html.Node, writer io.Writer)

	f = func(n *html.Node, w io.Writer) {

		if n.Type == html.ElementNode && n.Data == "head" {

			if h.options.AppendJS {

				for _, js := range h.options.JS {

					script_type := html.Attribute{"", "type", "text/javascript"}
					script_src := html.Attribute{"", "src", js}

					script := html.Node{
						Type:      html.ElementNode,
						DataAtom:  atom.Script,
						Data:      "script",
						Namespace: "",
						Attr:      []html.Attribute{script_type, script_src},
					}

					n.AppendChild(&script)
				}

			}

			if h.options.AppendCSS {

				for _, css := range h.options.CSS {
					link_type := html.Attribute{"", "type", "text/css"}
					link_rel := html.Attribute{"", "rel", "stylesheet"}
					link_href := html.Attribute{"", "href", css}

					link := html.Node{
						Type:      html.ElementNode,
						DataAtom:  atom.Link,
						Data:      "link",
						Namespace: "",
						Attr:      []html.Attribute{link_type, link_rel, link_href},
					}

					n.AppendChild(&link)
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "body" {

			if h.options.AppendAPIKey {
				api_key_ns := ""
				api_key_key := "data-mapzen-api-key"
				api_key_value := h.options.APIKey

				api_key_attr := html.Attribute{api_key_ns, api_key_key, api_key_value}
				n.Attr = append(n.Attr, api_key_attr)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, w)
		}
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	f(doc, wr)

	err = html.Render(wr, doc)

	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	wr.Flush()
	
	for k, v := range rec.Header() {
		rsp.Header()[k] = v
	}

	rsp.WriteHeader(200)

	data := buf.Bytes()
	clen := len(data)

	req.Header.Set("Content-Length", strconv.Itoa(clen))
	rsp.Write(data)
}
