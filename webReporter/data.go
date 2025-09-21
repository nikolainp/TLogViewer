package webreporter

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

func (obj *WebReporter) dataSource(w http.ResponseWriter, req *http.Request) {
	var js string

	processId := req.PathValue("source")

	toDataRows := func(data map[string]process) []string {
		//{"c":[{"v":"Mushrooms"},{"v":"3333"},{"v":"3333"},{"v":"333"},{"v":"Date(2010,10,6,12,13,14)"},{"v":"Date(2010,10,6,15,16,17)"}]}

		rows := make([]string, 0, len(data))

		for i := range data {
			rows = append(rows, fmt.Sprintf(
				`{"c":[{"v":"%s"},{"v":"%s"},{"v":"%s"},{"v":"%s"},{"v":"Date(%s)"},{"v":"Date(%s)"}]}`,
				template.JSEscapeString(data[i].Name),
				data[i].ServerName,
				data[i].IP,
				data[i].Port,
				data[i].FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				data[i].LastEventTime.Format("2006, 01, 02, 15, 04, 05"),
			))
		}

		return rows
	}

	if processId == "processes.json" {

		js = `
	{
	"cols": [
			{"id":"","label":"Process","type":"string"},
			{"id":"","label":"Server","type":"string"},
			{"id":"","label":"IP","type":"string"},
			{"id":"","label":"Port","type":"string"},
			{"id":"","label":"First event","type":"datetime"},
			{"id":"","label":"Last event","type":"datetime"}
		],
	"rows": [
			%s
		]
 	}
`
		js = fmt.Sprintf(js, strings.Join(toDataRows(obj.getProcesses()), ","))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(js))
}
