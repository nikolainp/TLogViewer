package webreporter

import (
	"fmt"
	"net/http"
)

type Storage interface {
}

type WebReporter struct {
	storage Storage
	srv     http.Server
}

func New(storage Storage) *WebReporter {
	obj := new(WebReporter)

	obj.storage = storage
	obj.srv = http.Server{
		Handler: obj.getHandlers(),
		Addr:    ":8090",
	}

	return obj
}

func (obj *WebReporter) Start() {
	obj.srv.ListenAndServe()
}

///////////////////////////////////////////////////////////////////////////////

func (obj *WebReporter) getHandlers() *http.ServeMux {
	sm := http.NewServeMux()
	sm.HandleFunc("/", obj.hello)
	sm.HandleFunc("/headers", obj.headers)
	return sm
}

func (obj *WebReporter) hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func (obj *WebReporter) headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
