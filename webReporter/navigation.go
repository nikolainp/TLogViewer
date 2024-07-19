package webreporter

import (
	"strings"
	"text/template"
)

type navigation struct {
}

func (obj *navigation) getContent() string {
	w := new(strings.Builder)
	sample, err := template.New("navigation").Parse(navagationTemplate)
	checkErr(err)

	data := struct{}{}

	err = sample.Execute(w, data)
	checkErr(err)

	return w.String()
}

const navagationTemplate = `
	<style  type="text/css">
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
			width: 5em;
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
	</style>

	<nav class="menu">	
		<ul class="nav">
			<li><a href="/">главная</a></li>
			<li><a href="/processes">процессы</a></li>
			<li><a href="/performance">производительность</a></li>
		</ul>
	</nav>
`
