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

func (obj *navigation) getSubMenu(url string, menuItems map[string]string) string {
	w := new(strings.Builder)
	sample, err := template.New("navigation2").Parse(navigationSubMenuTemplate)
	checkErr(err)

	// getMenuItems := func(menuItems map[string]string) (res []struct{ Id, Name string }) {
	// 	for item, value := range menuItems {
	// 		res = append(res, struct {
	// 			Id   string
	// 			Name string
	// 		}{Id: item, Name: value})
	// 	}
	// 	return
	// }

	data := struct {
		URL       string
		MenuItems map[string]string
	}{
		URL:       url,
		MenuItems: menuItems,
	}

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

const navigationSubMenuTemplate = `
	{{ $url := .URL }}
	<nav class="menu">	
		<ul class="nav" style="display: inline-grid;">
		{{ range $index, $item := .MenuItems }}
			<li><a style="width: 150px; overflow-wrap: anywhere;" 
				href="{{$url}}/{{$index}}">{{$item}}</a></li>
		{{end}}
		</ul>
	</nav>
`
