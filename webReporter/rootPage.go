package webreporter

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/nikolainp/TLogViewer/storage"
)

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []storage.Process) []string {

		rows := make([]string, len(data))

		for i := range data {
			rows[i] = fmt.Sprintf("['%s', '%s', '%s', new Date(%s), new Date(%s), null, 100, null]",
				data[i].Process, data[i].Process, data[i].Catalog,
				data[i].FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				data[i].LastEventTime.Format("2006, 01, 02, 15, 04, 05"))
		}

		return rows
	}

	dataGraph, err := template.New("dataGraph").Parse(rootPageTemplate)
	checkErr(err)

	processes, err := obj.storage.SelectAllProcesses()
	checkErr(err)

	data := struct {
		Processes []string
	}{
		Processes: toDataRows(processes),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)

}

const rootPageTemplate = `
<html>
<head>
<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
<script type="text/javascript">

  // Load the Visualization API and the controls package.
  google.charts.load('current', {'packages':['corechart', 'controls']});

  // Set a callback to run when the Google Visualization API is loaded.
  google.charts.setOnLoadCallback(drawDashboard);

  function drawDashboard() {

      var data = new google.visualization.DataTable();
      data.addColumn('string', 'Task ID');
      data.addColumn('string', 'Task Name');
      data.addColumn('string', 'Resource');
      data.addColumn('date', 'Start Date');
      data.addColumn('date', 'End Date');
      data.addColumn('number', 'Duration');
      data.addColumn('number', 'Percent Complete');
      data.addColumn('string', 'Dependencies');

      data.addRows([
		{{- range $process := .Processes -}}
			{{$process}},
	   	{{- end}}
	  ]);

	  var dashboard = new google.visualization.Dashboard(
		document.getElementById('timeline_div'));


	  var serverFilter = new google.visualization.ControlWrapper({
		'controlType': 'CategoryFilter',
		'containerId': 'serverfilter_div',
		'options': {
		  'filterColumnLabel': 'Resource',
		  'ui': {label: 'Server', labelSeparator: ':'},
		}
	  });

	  var trackHeight = 60;
	  var ganttChart = new google.visualization.ChartWrapper({
		'chartType': 'Gantt',
		'containerId': 'gantt_div',
		'dataTable': data,
		'options': {
			'height': data.getNumberOfRows() * trackHeight + trackHeight,
			'percentEnabled': false,
		}
	  });

	  dashboard.bind(serverFilter, ganttChart);
	  dashboard.draw(data);
    }
  </script>
</head>
<body>
	<div id="timeline_div">
		<div id="serverfilter_div"></div>
  		<div id="gantt_div"></div>
	</div>
</body>
</html>
`
