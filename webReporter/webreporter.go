package webreporter

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/nikolainp/TLogViewer/storage"
)

type WebReporter struct {
	storage *storage.Storage
	srv     http.Server
}

func New(storage *storage.Storage) *WebReporter {
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
	sm.HandleFunc("/", obj.rootPage)
	sm.HandleFunc("/headers", obj.headers)
	return sm
}

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	dataGraph, err := template.New("dataGraph").Parse(rootPageTemplate)
	checkErr(err)

	processes, err := obj.storage.SelectAllProcesses()
	checkErr(err)

	data := struct {
		Processes []storage.Process
	}{
		Processes: processes,
	}

	err = dataGraph.Execute(w, data)
	checkErr(err)

}

func (obj *WebReporter) headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

var checkErr = func(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const rootPageTemplate = `
{{range $process := .Processes -}}
{{$process.Name}}
{{end}}
`
