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

	.dropdown {
		position: relative;
		display: inline-block;
	}
	
	.dropdown-content {
		display: none;
		position: absolute;
		background-color: #f9f9f9;
		min-width: 160px;
		box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
		padding: 12px 16px;
		z-index: 1;
	}
	
	.dropdown:hover .dropdown-content {
		display: block;
	}

`
