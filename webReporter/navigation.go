package webreporter

import (
	"strings"
	"text/template"
)

type navigation struct {
}

func (obj *navigation) getMainMenu() string {
	w := new(strings.Builder)
	sample, err := template.New("navigation").Parse(navigationMainMenuTemplate)
	checkErr(err)

	data := struct{}{}

	err = sample.Execute(w, data)
	checkErr(err)

	return w.String()
}

const navigationMainMenuTemplate = `

	<nav class="menu">	
		<ul class="nav">
			<li><a href="/">главная</a></li>
			<li><a href="/processes">процессы</a></li>
			<li><a href="/performance">производительность</a></li> 
		</ul>
	</nav>
`
