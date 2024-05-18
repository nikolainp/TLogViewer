package webreporter

import (
	"net/http"
	"text/template"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	dataGraph, err := template.New("rootPage").Parse(processesTemplate)
	checkErr(err)

	title, err := obj.storage.SelectDetails()
	checkErr(err)

	data := struct {
		Title string
	}{
		Title: title,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)
}
