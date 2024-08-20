package webreporter

import (
	"net/http"
	"text/template"
)

func (obj *WebReporter) static(w http.ResponseWriter, req *http.Request) {
	dataGraph, err := template.New("style.css").Parse(styleCSS)
	checkErr(err)

	data := struct{}{}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)

}

const styleCSS = `
	ul.nav{
		margin-left: 0px;
		padding-left: 0px;
		list-style: none;
	}
	.nav li { 
		display: inline; 
	}
	ul.nav a {
		display: inline-block;
		padding: 5px;
		background-color: #f4f4f4;
		border: 1px dashed #333;
		text-decoration: none;
		color: #333;
		text-align: center;
	}
	ul.nav a:hover{
		background-color: #333;
		color: #f4f4f4;
	}
`
