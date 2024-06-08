package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	dataGraph, err := template.New("rootPage").Parse(rootPageTemplate)
	checkErr(err)

	data := struct {
		Details rootDetails
	}{
		Details: obj.getRootDetails(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

type rootDetails struct {
	Title, Version                                string
	ProcessingSize, ProcessingSpeed               string
	ProcessingTime, FirstEventTime, LastEventTime string
}

func (obj *WebReporter) getRootDetails() (data rootDetails) {

	var processingTime time.Time
	var firstEventTime, lastEventTime time.Time

	details := obj.storage.SelectAll("details", "")
	details.Next(
		&data.Title, &data.Version,
		&data.ProcessingSize, &data.ProcessingSpeed,
		&processingTime,
		&firstEventTime, &lastEventTime)

	data.ProcessingTime = processingTime.Format("2006-01-02 15:04:05")
	data.FirstEventTime = firstEventTime.Format("2006-01-02 15:04:05")
	data.LastEventTime = lastEventTime.Format("2006-01-02 15:04:05")

	details.Next()

	return
}

const rootPageTemplate = `
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">

  <title>{{.Details.Title}}</title>


  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">

    // Load the Visualization API and the controls package.
    google.charts.load('current', {'packages':['corechart', 'controls']});

    // Set a callback to run when the Google Visualization API is loaded.
    google.charts.setOnLoadCallback(drawDashboard);

    function drawDashboard() {

	}

	</script>
</head>
<body>
	<div>
		<h2>Источник данных: {{.Details.Title}}</h2>
		<h3>дата обработки: {{.Details.ProcessingTime}} (версия {{.Details.Version}})</br>
			размер данных: {{.Details.ProcessingSize}}</br>
			скорость обработки: {{.Details.ProcessingSpeed}}/сек.</h3>
	</div>
	<hr>
	<div>
		<h3>данные собирались с {{.Details.FirstEventTime}} по {{.Details.LastEventTime}}</h3>
	</div>
	<hr>
	<a href="/processes">процессы</a>
	<hr>

</body>
</html>
`

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
