package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []process) []string {

		rows := make([]string, len(data))

		for i := range data {
			rows[i] = fmt.Sprintf("['%s', '%s', '%s', '%s', new Date(%s), new Date(%s)]",
				template.JSEscapeString(data[i].Name),
				data[i].ServerName,
				data[i].IP,
				data[i].Port,
				data[i].FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				data[i].LastEventTime.Format("2006, 01, 02, 15, 04, 05"))
		}

		return rows
	}

	dataGraph, err := template.New("rootPage").Parse(rootPageTemplate)
	checkErr(err)

	details := obj.getRootDetails()

	data := struct {
		Title, Version                 string
		ProcessingSize, ProcessingTime string
		ProcessingSpeed                string
		FirstEventTime, LastEventTime  string
		DataFilter                     string
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
		Processes:       toDataRows(obj.getProcesses()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

type rootDetails struct {
	Title, Version                                string
	ProcessingSize, ProcessingSpeed               int64
	ProcessingTime, FirstEventTime, LastEventTime time.Time
}

func (obj *WebReporter) getRootDetails() (data rootDetails) {

	details := obj.storage.SelectAll("details", "")
	details.Next(
		&data.Title, &data.Version,
		&data.ProcessingSize, &data.ProcessingSpeed,
		&data.ProcessingTime,
		&data.FirstEventTime, &data.LastEventTime)

	details.Next()

	return
}

const rootPageTemplate = `
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">

  <title>{{.Title}}</title>


  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">

    // Load the Visualization API and the controls package.
	google.charts.load('current', {'packages':['table']});

    // Set a callback to run when the Google Visualization API is loaded.
    google.charts.setOnLoadCallback(drawDashboard);

    function drawDashboard() {
	    var data = new google.visualization.DataTable();
        data.addColumn('string', 'Process');
        data.addColumn('string', 'Server');
        data.addColumn('string', 'IP');
        data.addColumn('string', 'Port');
        data.addColumn('datetime', 'First event');
        data.addColumn('datetime', 'Last event');

        data.addRows([
		{{- range $process := .Processes -}}
			{{$process}},
	   	{{- end}}
        ]);

        var table = new google.visualization.Table(document.getElementById('table_div'));

		var dateFormat = new google.visualization.DateFormat({pattern: "YYYY-MM-dd HH:mm:ss"});
		dateFormat.format(data, 4);
		dateFormat.format(data, 5);

        table.draw(data, {showRowNumber: false, width: '100%'});
	}

	</script>
</head>
<body>
	<div>
		<h2>Источник данных: {{.Title}}</h2>
		<h3>дата обработки: {{.ProcessingTime}} (версия {{.Version}})</br>
			размер данных: {{.ProcessingSize}},
			скорость обработки: {{.ProcessingSpeed}}/сек.</h3>
	</div>
	<hr>
	{{.DataFilter}}
	<hr>
	<a href="/processes">процессы</a>
	<hr>
	<div id="table_div"></div>

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
