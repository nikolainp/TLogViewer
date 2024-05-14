package webreporter

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/nikolainp/TLogViewer/storage"
)

type WebReporter struct {
	storage *storage.Storage
	srv     http.Server
}

func New(storage *storage.Storage) *WebReporter {
	obj := new(WebReporter)

	obj.storage = storage
	obj.srv = http.Server{
		Handler: obj.getHandlers(),
		Addr:    ":8090",
	}

	return obj
}

func (obj *WebReporter) Start() {
	obj.srv.ListenAndServe()
}

///////////////////////////////////////////////////////////////////////////////

func (obj *WebReporter) getHandlers() *http.ServeMux {
	sm := http.NewServeMux()
	sm.HandleFunc("/", obj.rootPage)
	sm.HandleFunc("/headers", obj.headers)
	return sm
}

func (obj *WebReporter) rootPage(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []storage.Process) []string {
		// ['Research', 'Find sources', null,
		//  new Date(2015, 0, 1), new Date(2015, 0, 5), null,  100,  null],
		// ['Write', 'Write paper', 'write',
		//  null, new Date(2015, 0, 9), daysToMilliseconds(3), 25, 'Research,Outline'],
		// ['Cite', 'Create bibliography', 'write',
		//  null, new Date(2015, 0, 7), daysToMilliseconds(1), 20, 'Research'],
		// ['Complete', 'Hand in paper', 'complete',
		//  null, new Date(2015, 0, 10), daysToMilliseconds(1), 0, 'Cite,Write'],
		// ['Outline', 'Outline paper', 'write',
		//  null, new Date(2015, 0, 6), daysToMilliseconds(1), 100, 'Research']

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

func (obj *WebReporter) headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

var checkErr = func(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

  // Callback that creates and populates a data table,
  // instantiates a dashboard, a range slider and a pie chart,
  // passes in the data and draws it.
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
      //chart.draw(data, options);
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
