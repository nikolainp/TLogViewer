package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) processes(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []process) []string {

		rows := make([]string, len(data))

		for i := range data {
			firstEventTime := obj.filter.getStartTime(data[i].FirstEventTime)
			lastEventTime := obj.filter.getFinishTime(data[i].LastEventTime)

			rows[i] = fmt.Sprintf("['%s', '%s', '%s', new Date(%s), new Date(%s), null, 100, null]",
				data[i].ProcessID,
				template.JSEscapeString(data[i].Name),
				data[i].Catalog,
				firstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				lastEventTime.Format("2006, 01, 02, 15, 04, 05"))
		}

		return rows
	}

	dataGraph, err := template.New("processes").Parse(processesTemplate)
	checkErr(err)

	data := struct {
		Title      string
		DataFilter string
		Navigation string
		Processes  []string
	}{
		Title:      obj.title,
		DataFilter: obj.filter.getContent(req.URL.String()),
		Navigation: obj.navigator.getContent(),
		Processes:  toDataRows(obj.getProcesses()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
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
}

func (obj *WebReporter) getProcesses() (data []process) {

	var elem process

	data = make([]process, 0)

	details := obj.storage.SelectQuery("processes")
	details.SetTimeFilter(obj.filter.getData())
	for details.Next(
		&elem.Name, &elem.Catalog, &elem.Process,
		&elem.ProcessID, &elem.ProcessType,
		&elem.Pid, &elem.Port, &elem.UID,
		&elem.ServerName, &elem.IP,
		&elem.FirstEventTime, &elem.LastEventTime) {

		data = append(data, elem)
	}

	return
}

const processesTemplate = `
<html>
<head>

  <title>{{.Title}} | Processes</title>

  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">

    // Load the Visualization API and the controls package.
    google.charts.load('current', {'packages':['corechart', 'controls'], 'language': 'ru'});

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
			'label': 'Время жизни процессов',
			'height': data.getNumberOfRows() * trackHeight + trackHeight,
			'percentEnabled': false,
		}
	  });

	  dashboard.bind(serverFilter, ganttChart);
	  dashboard.draw(data);

	  google.visualization.events.addListener(ganttChart, 'select', selectHandler);

	  function selectHandler(e) {
		var selection = ganttChart.getChart().getSelection();
		if (selection.length != 1) {
			return;
		}
		var item = selection[0];
		if (item.row == null) {
			return;
		}
		var str = ganttChart.getDataTable().getValue(item.row, 0);
		
		location.replace('/performance?processId=' + str)
	  }

	}

	</script>
</head>
<body>
	{{.DataFilter}}
	{{.Navigation}}
	<div id="timeline_div">
		<div id="serverfilter_div" style="vertical-align: middle;"></div>
  		<div id="gantt_div" style="width: 100%; height: calc(100% - 30px);">></div>
	</div>
</body>
</html>
`
