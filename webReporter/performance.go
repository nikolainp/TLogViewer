package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) performance(w http.ResponseWriter, req *http.Request) {

	toDataRows := func(data []process) []string {

		rows := make([]string, len(data))

		for i := range data {
			firstEventTime := obj.filter.getStartTime(data[i].FirstEventTime)
			lastEventTime := obj.filter.getFinishTime(data[i].LastEventTime)

			rows[i] = fmt.Sprintf("['%s', '%s', '%s', new Date(%s), new Date(%s), null, 100, null]",
				data[i].Process,
				template.JSEscapeString(data[i].Name),
				data[i].Catalog,
				firstEventTime.Format("2006, 01, 02, 15, 04, 05"),
				lastEventTime.Format("2006, 01, 02, 15, 04, 05"))
		}

		return rows
	}

	dataGraph, err := template.New("performance").Parse(performanceTemplate)
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

type performance struct {
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

func (obj *WebReporter) getPerformanceStatistics() {
	details := obj.storage.SelectQuery("processesPerformance", "")
	details.SetFilter(obj.filter.getData())
	details.SetGroup("processID")
	for details.Next() {

	}
}

func (obj *WebReporter) getPerformance() (data []process) {

	var elem process

	data = make([]process, 0)

	details := obj.storage.SelectQuery("processesPerformance", "")
	details.SetFilter(obj.filter.getData())
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

const performanceTemplate = `
<html>
<head>

  <title>{{.Title}} | Performance</title>

  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">
      google.charts.load('current', {'packages':['annotationchart']});
      google.charts.setOnLoadCallback(drawChart);

      let chart;

      function drawChart() {
        var columns = new google.visualization.DataTable();
        columns.addColumn('string', 'Title');
        columns.addColumn('number', 'Minimum');
        columns.addColumn('number', 'Maximum');
        columns.addColumn('number', 'Average');
        columns.addColumn('boolean', 'Show');

        columns.addRows([
        {{range $i, $column := .Columns -}}
          ['{{$column.Name}}', {{$column.Minimum}}, {{$column.Maximum}}, {{$column.Average}}, true],
        {{end}}
        ]);


        var data = new google.visualization.DataTable();
        data.addColumn('date', 'Date');
        {{range $column := .Columns -}}
        data.addColumn('number', '{{$column.Name}}');
        {{end}}
		    // [new Date(2314, 2, 16), 24045, 12374],
        
        data.addRows([
			{{- range $index, $dataRow := .DataRows -}}
			  {{- if (eq $index 0)}}
           {{$dataRow}}
        {{- else}}
          ,{{$dataRow}}
        {{- end}}
			{{- end}}
        ]);

        table = new google.visualization.Table(document.getElementById('table_div'));
        chart = new google.visualization.AnnotationChart(document.getElementById('chart_div'));

        chart.draw(data, {displayAnnotations: true});
        setColumnColors(columns);
        drawTable(table, columns)

        google.visualization.events.addListener(table, 'select', function() {
          var row = table.getSelection()[0].row;
          var value = columns.getValue(row, 4);

          value = !value;
          columns.setValue(row, 4, value);
          if (value) {
              chart.showDataColumns(row)
          } else {
              chart.hideDataColumns(row)
          }

          drawTable(table, columns);
        });

      }

      function setColumnColors(table) {
        var legends = document.getElementsByClassName("legend-dot");

        for (let i = 0; i < legends.length; i++) {
            table.setProperties(i, 0, {'style': 'color: ' + legends[i].style.backgroundColor +'; white-space: nowrap;'});
            table.setProperties(i, 4, {'style': 'color: ' + legends[i].style.backgroundColor +';'});
        }
      }

      function drawTable(table, data) {
        table.draw(data, {showRowNumber: true, allowHtml: true, width: '100%', height: '100%'});
      }

      </script>
	</script>
</head>
<body>
	{{.DataFilter}}
	{{.Navigation}}

	<div class="dropdown" style='height: 30px;'>
      <span>{{.Title}}</span>
      <div class="dropdown-content">
        <div id='table_div'></div>
      </div>
    </div>

    <div id='chart_div' style='width: 100%; height: calc(100% - 30px);'></div>
</body>
</html>
`
