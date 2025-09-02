package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data map[string]process) []string {

		rows := make([]string, 0, len(data))

		for i := range data {
			rows = append(rows, fmt.Sprintf("['%s', '%s', '%s', '%s', new Date(%s), new Date(%s)]",
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
		Processes:       toDataRows(obj.getProcesses()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "rootPage.gohtml", data)
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
