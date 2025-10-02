package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		obj.logger.Printf("NotFound: %s %s", req.Method, req.URL.Path)
		http.NotFound(w, req)
		return
	}

	details := obj.getRootDetails()

	data := struct {
		Title, Version                 string
		ProcessingSize, ProcessingTime string
		ProcessingSpeed                string
		FirstEventTime, LastEventTime  string
		DataFilter                     string
		Navigation                     string
		Processes                      []string
	}{
		Title:           obj.title,
		Version:         details.Version,
		ProcessingSize:  byteCount(details.ProcessingSize),
		ProcessingTime:  details.ProcessingTime.Format("2006-01-02 15:04:05"),
		ProcessingSpeed: byteCount(details.ProcessingSpeed),
		FirstEventTime:  details.FirstEventTime.Format("2006-01-02 15:04:05"),
		LastEventTime:   details.LastEventTime.Format("2006-01-02 15:04:05"),
		DataFilter:      obj.filter.getContent(req.URL.String()),
		Navigation:      obj.navigator.getMainMenu(),
		//		Processes:       toDataRows(obj.getProcesses()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "rootPage.html", data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

type rootDetails struct {
	Title, Version                                string
	ProcessingSize, ProcessingSpeed               int64
	ProcessingTime, FirstEventTime, LastEventTime time.Time
}

func (obj *WebReporter) getRootDetails() (data rootDetails) {

	details := obj.storage.SelectQuery("details")
	details.Next(
		&data.Title, &data.Version,
		&data.ProcessingSize, &data.ProcessingSpeed,
		&data.ProcessingTime,
		&data.FirstEventTime, &data.LastEventTime)

	details.Next()

	return
}

func (obj *WebReporter) getProcesses() (data dataSource) {

	var elem process

	data.columns = make([]string, 6)
	data.columns[0] = `{"id":"","label":"Process","type":"string"}`
	data.columns[1] = `{"id":"","label":"Server","type":"string"}`
	data.columns[2] = `{"id":"","label":"IP","type":"string"}`
	data.columns[3] = `{"id":"","label":"Port","type":"string"}`
	data.columns[4] = `{"id":"","label":"First event","type":"datetime"}`
	data.columns[5] = `{"id":"","label":"Last event","type":"datetime"}`

	data.rows = make([]string, 0)

	details := obj.storage.SelectQuery("processes")
	details.SetTimeFilter(obj.filter.getData())
	details.SetOrder("Name")

	for details.Next(
		&elem.Name, &elem.Catalog, &elem.Process,
		&elem.ProcessID, &elem.ProcessType,
		&elem.Pid, &elem.Port, &elem.UID,
		&elem.ServerName, &elem.IP,
		&elem.FirstEventTime, &elem.LastEventTime) {

		data.rows = append(data.rows, fmt.Sprintf(
			`{"c":[{"v":"%s"},{"v":"%s"},{"v":"%s"},{"v":"%s"},{"v":"Date(%s)"},{"v":"Date(%s)"}]}`,
			template.JSEscapeString(elem.Name),
			elem.ServerName, elem.IP, elem.Port,
			elem.FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
			elem.LastEventTime.Format("2006, 01, 02, 15, 04, 05"),
		))
	}

	return
}

func byteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%db", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}
