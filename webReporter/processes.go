package webreporter

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/nikolainp/TLogViewer/storage"
)

func (obj *WebReporter) processes(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []storage.Process) []string {

		rows := make([]string, len(data))

		for i := range data {
			rows[i] = fmt.Sprintf("['%s', '%s', '%s', new Date(%s), new Date(%s), null, 100, null]",
				data[i].Process,
				template.JSEscapeString(data[i].Name),
				data[i].Catalog,
				data[i].FirstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				data[i].LastEventTime.Format("2006, 01, 02, 15, 04, 05"))
		}

		return rows
	}

	dataGraph, err := template.New("processes").Parse(processesTemplate)
	checkErr(err)

	title, err := obj.storage.SelectDetails()
	checkErr(err)

	processes, err := obj.storage.SelectAllProcesses()
	checkErr(err)

	data := struct {
		Title     string
		Processes []string
	}{
		Title:     title,
		Processes: toDataRows(processes),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)

}

const processesTemplate = `
<html>
<head>

  <title>{{.Title}} | Processes</title>

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

	  var trackHeight = 50;
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

	  google.visualization.events.addListener(ganttChart, 'select', selectHandler);

	  function selectHandler(e) {
		alert('A table row was selected');
		var selection = ganttChart.getChart().getSelection();
		var message = '';
		for (var i = 0; i < selection.length; i++) {
		  var item = selection[i];
		  if (item.row != null && item.column != null) {
			var str = data.getFormattedValue(item.row, item.column);
			message += '{row:' + item.row + ',column:' + item.column + '} = ' + str + '\n';
		  } else if (item.row != null) {
			var str = data.getFormattedValue(item.row, 0);
			message += '{row:' + item.row + ', column:none}; value (col 0) = ' + str + '\n';
		  } else if (item.column != null) {
			var str = data.getFormattedValue(0, item.column);
			message += '{row:none, column:' + item.column + '}; value (row 0) = ' + str + '\n';
		  }
		}
		if (message == '') {
		  message = 'nothing';
		}
		alert('You selected ' + message);
	  }

	}

	</script>
</head>
<body>
	<div id="timeline_div">
		<div id="serverfilter_div"></div>
  		<div id="gantt_div" style="width: 100%; height: calc(100% - 30px);">></div>
	</div>
</body>
</html>
`
