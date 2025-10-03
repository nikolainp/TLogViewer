package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) processes(w http.ResponseWriter, req *http.Request) {

	// toDataRows := func(data map[string]process) []string {

	// 	rows := make([]string, 0, len(data))

	// 	for i := range data {
	// 		firstEventTime := obj.filter.getStartTime(data[i].FirstEventTime)
	// 		lastEventTime := obj.filter.getFinishTime(data[i].LastEventTime)

	// 		rows = append(rows, fmt.Sprintf("['%s', '%s', '%s', new Date(%s), new Date(%s), null, 100, null]",
	// 			data[i].ProcessID,
	// 			template.JSEscapeString(data[i].Name),
	// 			data[i].Catalog,
	// 			firstEventTime.Format("2006, 01, 02, 15, 04, 05"),
	// 			lastEventTime.Format("2006, 01, 02, 15, 04, 05"),
	// 		))
	// 	}

	// 	return rows
	// }

	data := struct {
		Title      string
		DataFilter string
		Navigation string
		//		Processes  []string
	}{
		Title:      obj.title,
		DataFilter: obj.filter.getContent(req.URL.String()),
		Navigation: obj.navigator.getMainMenu(),
		//		Processes:  toDataRows(obj.getProcessesLiveTime()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "processes.html", data)
	checkErr(err)

}

type process struct {
	Name           string
	Catalog        string
	Process        string
	ProcessID      string
	ProcessType    string
	Pid            string
	Port           string
	UID            string
	ServerName     string
	IP             string
	FirstEventTime time.Time
	LastEventTime  time.Time

	order int
}

func (obj *WebReporter) getProcessesLiveTime() (data dataSource) {

	var elem process

	data.columns = make([]string, 8)
	data.columns[0] = `{"id":"","label":"Task ID","type":"string"}`
	data.columns[1] = `{"id":"","label":"Task Name","type":"string"}`
	data.columns[2] = `{"id":"","label":"Resource","type":"string"}`
	data.columns[3] = `{"id":"","label":"Start Date","type":"date"}`
	data.columns[4] = `{"id":"","label":"End Date","type":"date"}`
	data.columns[5] = `{"id":"","label":"Duration","type":"number"}`
	data.columns[6] = `{"id":"","label":"Percent Complete","type":"number"}`
	data.columns[7] = `{"id":"","label":"Dependencies","type":"string"}`
	data.rows = make([]string, 0)

	details := obj.storage.SelectQuery("processes")
	details.SetTimeFilter(obj.filter.getData())
	details.SetOrder("Name")

	orderID := 0
	for details.Next(
		&elem.Name, &elem.Catalog, &elem.Process,
		&elem.ProcessID, &elem.ProcessType,
		&elem.Pid, &elem.Port, &elem.UID,
		&elem.ServerName, &elem.IP,
		&elem.FirstEventTime, &elem.LastEventTime) {

		elem.order = orderID
		///		data[elem.ProcessID] = elem

		data.rows = append(data.rows, fmt.Sprintf(
			`{"c":[{"v":"%s"},{"v":"%s"},{"v":"%s"},{"v":"Date(%s)"},{"v":"Date(%s)"},{"v":null},{"v":100},{"v":null}]}`,
			elem.ProcessID,
			template.JSEscapeString(elem.Name),
			elem.Catalog,
			elem.FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
			elem.LastEventTime.Format("2006, 01, 02, 15, 04, 05"),
		))

		orderID++
	}

	return
}
