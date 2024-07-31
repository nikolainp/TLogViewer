package webreporter

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func (obj *WebReporter) performance(w http.ResponseWriter, req *http.Request) {

	processId := req.URL.Query().Get("processId")

	toProcessDesc := func(data process) string {
		return fmt.Sprintf("%s (%s, %s:%s, start: %s - stop: %s)",
			data.Process,
			data.Name, data.IP, data.Port,
			data.FirstEventTime.Format("2006-01-02 15:04:05"),
			data.LastEventTime.Format("2006-01-02 15:04:05"),
		)
	}

	dataGraph, err := template.New("performance").Parse(performanceTemplate)
	checkErr(err)

	data := struct {
		Title      string
		DataFilter string
		Navigation string
		Process    string
		Columns    []dataColumn
		DataRows   []string
	}{
		Title:      obj.title,
		DataFilter: obj.filter.getContent(req.URL.String()),
		Navigation: obj.navigator.getContent(),
		Process:    toProcessDesc(obj.getProcess(processId)),
		Columns:    obj.getPerformanceStatistics(processId),
		DataRows:   obj.getPerformance(processId),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = dataGraph.Execute(w, data)
	checkErr(err)

}

type dataColumn struct {
	Name                      string
	Minimum, Average, Maximum float64
}

func (obj *WebReporter) getProcess(processId string) (data process) {

	details := obj.storage.SelectQuery("processes")
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processID = ?", processId)
	details.Next(
		&data.Name, &data.Catalog, &data.Process,
		&data.ProcessID, &data.ProcessType,
		&data.Pid, &data.Port, &data.UID,
		&data.ServerName, &data.IP,
		&data.FirstEventTime, &data.LastEventTime)
	details.Next()

	return
}

func (obj *WebReporter) getPerformanceStatistics(processId string) (data []dataColumn) {

	var processID string
	var MIN_cpu, AVG_cpu, MAX_cpu float64
	var MIN_queue_length, AVG_queue_length, MAX_queue_length float64
	var MIN_queue_lengthByCpu, AVG_queue_lengthByCpu, MAX_queue_lengthByCpu float64
	var MIN_memory_performance, AVG_memory_performance, MAX_memory_performance float64
	var MIN_disk_performance, AVG_disk_performance, MAX_disk_performance float64
	var MIN_response_time, AVG_response_time, MAX_response_time float64
	var MIN_average_response_time, AVG_average_response_time, MAX_average_response_time float64

	details := obj.storage.SelectQuery("processesPerformance", "processID",
		"MIN(cpu)", "AVG(cpu)", "MAX(cpu)",
		"MIN(queue_length)", "AVG(queue_length)", "MAX(queue_length)",
		"MIN(queue_lengthByCpu)", "AVG(queue_lengthByCpu)", "MAX(queue_lengthByCpu)",
		"MIN(memory_performance)", "AVG(memory_performance)", "MAX(memory_performance)",
		"MIN(disk_performance)", "AVG(disk_performance)", "MAX(disk_performance)",
		"MIN(response_time)", "AVG(response_time)", "MAX(response_time)",
		"MIN(average_response_time)", "AVG(average_response_time)", "MAX(average_response_time)",
	)
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processID = ?", processId)
	details.SetGroup("processID")
	details.Next(
		&processID,
		&MIN_cpu, &AVG_cpu, &MAX_cpu,
		&MIN_queue_length, &AVG_queue_length, &MAX_queue_length,
		&MIN_queue_lengthByCpu, &AVG_queue_lengthByCpu, &MAX_queue_lengthByCpu,
		&MIN_memory_performance, &AVG_memory_performance, &MAX_memory_performance,
		&MIN_disk_performance, &AVG_disk_performance, &MAX_disk_performance,
		&MIN_response_time, &AVG_response_time, &MAX_response_time,
		&MIN_average_response_time, &AVG_average_response_time, &MAX_average_response_time,
	)
	details.Next()

	data = make([]dataColumn, 7)
	data[0] = dataColumn{Name: "cpu", Minimum: MIN_cpu, Average: AVG_cpu, Maximum: MAX_cpu}
	data[1] = dataColumn{Name: "queue_length", Minimum: MIN_queue_length, Average: AVG_queue_length, Maximum: MAX_queue_length}
	data[2] = dataColumn{Name: "queue_lengthByCpu", Minimum: MIN_queue_lengthByCpu, Average: AVG_queue_lengthByCpu, Maximum: MAX_queue_lengthByCpu}
	data[3] = dataColumn{Name: "memory_performance", Minimum: MIN_memory_performance, Average: AVG_memory_performance, Maximum: MAX_memory_performance}
	data[4] = dataColumn{Name: "disk_performance", Minimum: MIN_disk_performance, Average: AVG_disk_performance, Maximum: MAX_disk_performance}
	data[5] = dataColumn{Name: "response_time", Minimum: MIN_response_time, Average: AVG_response_time, Maximum: MAX_response_time}
	data[6] = dataColumn{Name: "average_response_time", Minimum: MIN_average_response_time, Average: AVG_average_response_time, Maximum: MAX_average_response_time}

	return
}

func (obj *WebReporter) getPerformance(processId string) (data []string) {

	var eventTime string
	var cpu, queue_length, queue_lengthByCpu float64
	var memory_performance, disk_performance float64
	var response_time, average_response_time float64

	data = make([]string, 0)

	details := obj.storage.SelectQuery("processesPerformance", "eventTime",
		"cpu", "queue_length", "queue_lengthByCpu",
		"memory_performance", "disk_performance",
		"response_time", "average_response_time")
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processID = ?", processId)
	for details.Next(&eventTime,
		&cpu, &queue_length, &queue_lengthByCpu,
		&memory_performance, &disk_performance,
		&response_time, &average_response_time) {

		eventTTime, err := time.ParseInLocation("2006-01-02 15:04:05", eventTime[:19], time.Local)
		checkErr(err)

		data = append(data, fmt.Sprintf("[new Date(%s), %g, %g, %g, %g, %g, %g, %g]",
			eventTTime.Format("2006, 01, 02, 15, 04, 05"),
			cpu, queue_length, queue_lengthByCpu,
			memory_performance, disk_performance,
			response_time, average_response_time,
		))
	}

	return
}

const performanceTemplate = `
<html>
<head>

  <title>{{.Title}} | Performance</title>

      <style>
        .dropdown {
          position: relative;
          display: inline-block;
        }
        
        .dropdown-content {
          display: none;
          position: absolute;
          background-color: #f9f9f9;
          min-width: 160px;
          box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
          padding: 12px 16px;
          z-index: 1;
        }
        
        .dropdown:hover .dropdown-content {
          display: block;
        }
    </style>


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
      <span>{{.Process}}</span>
      <div class="dropdown-content">
        <div id='table_div'></div>
      </div>
    </div>

    <div id='chart_div' style='width: 100%; height: calc(100% - 30px);'></div>
</body>
</html>
`
