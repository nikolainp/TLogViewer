package webreporter

import (
	"sort"
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

	getMenuItems := func(menuItems map[string]string) (res []struct{ Id, Name string }) {
		for item, value := range menuItems {
			res = append(res, struct {
				Id   string
				Name string
			}{Id: item, Name: value})
		}
		sort.Slice(res, func(i, j int) bool { return strings.Compare(res[i].Name, res[j].Name) < 0 })
		return
	}

	data := struct {
		URL       string
		MenuItems []struct{ Id, Name string }
	}{
		URL:       url,
		MenuItems: getMenuItems(menuItems),
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
		{{ range $item := .MenuItems }}
			<li><a style="width: 150px; overflow-wrap: anywhere;" 
				{{ if (eq $item.Id "") }}
				href="{{$url}}">{{$item.Name}}
				{{ else }}
				href="{{$url}}/{{$item.Id}}">{{$item.Name}}
				{{ end }}
				</a></li>
		{{end}}
		</ul>
	</nav>
`
