package webreporter

import (
	"fmt"
	"net/http"
	"time"
)

func (obj *WebReporter) performance(w http.ResponseWriter, req *http.Request) {

	urlString := "/performance"

	data := struct {
		Title      string
		DataFilter string
		Navigation string

		ProcessList string

		ProcessId string
		Process   string
		// Columns  []dataColumn
		// DataRows []string
	}{}

	toProcessList := func(data map[string]process) (res map[string]string) {
		res = make(map[string]string)
		res[""] = "Все"
		for i := range data {
			if len(data[i].Name) == 0 {
				res[i] = "process_" + data[i].Pid
			} else {
				res[i] = data[i].Name
			}
		}
		return
	}
	toProcessDesc := func(data process) string {
		return fmt.Sprintf("%s (%s, %s:%s, start: %s - stop: %s)",
			data.Process,
			data.Name, data.IP, data.Port,
			data.FirstEventTime.Format("2006-01-02 15:04:05"),
			data.LastEventTime.Format("2006-01-02 15:04:05"),
		)
	}

	processList := obj.getWorkProcesses()
	data.Title = obj.title
	data.DataFilter = obj.filter.getContent(req.URL.String())
	data.Navigation = obj.navigator.getMainMenu()
	data.ProcessList = obj.navigator.getSubMenu(urlString, toProcessList(processList))

	data.ProcessId = req.PathValue("id")
	if data.ProcessId == "" {
		// total data
		data.Process = "Производительность рабочих процессов"
	} else {
		// by process
		data.Process = toProcessDesc(obj.getWorkProcess(data.ProcessId))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "performance.html", data)
	checkErr(err)
}

type dataColumn struct {
	Name                      string
	Minimum, Average, Maximum float64
}

func (obj *WebReporter) getProcess(processId string) (data process) {

	details := obj.storage.SelectQuery("processes")
	//details.SetTimeFilter(obj.filter.getData())
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

func (obj *WebReporter) getWorkProcesses() (data map[string]process) {

	var elem process

	data = make(map[string]process, 0)

	details := obj.storage.SelectQuery("workProcesses")
	details.SetTimeFilter(obj.filter.getData())
	details.SetOrder("Name", "Pid")

	orderID := 0
	for details.Next(
		&elem.ProcessID, &elem.Name,
		&elem.Pid, &elem.Port, &elem.ServerName,
		&elem.FirstEventTime, &elem.LastEventTime) {

		elem.order = orderID
		data[elem.ProcessID] = elem
		orderID++
	}

	return
}

func (obj *WebReporter) getWorkProcess(processId string) (data process) {

	details := obj.storage.SelectQuery("workProcesses")
	//details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processWID = ?", processId)
	details.Next(
		&data.ProcessID, &data.Name,
		&data.Pid, &data.Port, &data.ServerName,
		&data.FirstEventTime, &data.LastEventTime)
	details.Next()

	return
}

func (obj *WebReporter) getPerformanceStatistics(processId string) (data dataSource) {
	if processId == "" {
		return obj.getTotalPerformanceStatistics()
	}
	return obj.getProcessPerformanceStatistics(processId)
}

func (obj *WebReporter) getPerformance(processId string) (data dataSource) {
	if processId == "" {
		return obj.getTotalPerformance()
	}
	return obj.getProcessPerformance(processId)
}

func (obj *WebReporter) getTotalPerformanceStatistics() (data dataSource) {

	var processID string
	var MIN_average_response_time, AVG_average_response_time, MAX_average_response_time float64

	data.columns = make([]string, 5)
	data.columns[0] = `{"id":"","label":"Process","type":"string"}`
	data.columns[1] = `{"id":"","label":"Minimun","type":"number"}`
	data.columns[2] = `{"id":"","label":"Maximum","type":"number"}`
	data.columns[3] = `{"id":"","label":"Average","type":"number"}`
	data.columns[4] = `{"id":"","label":"Show","type":"boolean"}`
	data.rows = make([]string, 0)

	details := obj.storage.SelectQuery("processesPerformance", "processWID",
		"MIN(average_response_time)", "AVG(average_response_time)", "MAX(average_response_time)",
	)
	details.SetTimeFilter(obj.filter.getData())
	details.SetGroup("processWID")
	details.SetOrder("processWID")

	for details.Next(
		&processID,
		&MIN_average_response_time, &AVG_average_response_time, &MAX_average_response_time,
	) {
		data.rows = append(data.rows, fmt.Sprintf(
			`{"c":[{"v":"%s"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`,
			processID,
			10000/MIN_average_response_time,
			10000/AVG_average_response_time,
			10000/MAX_average_response_time,
		))
	}

	return
}

func (obj *WebReporter) getTotalPerformance() (data dataSource) {

	var eventTime string
	var processID string
	var average_response_time float64

	details := obj.storage.SelectQuery("processesPerformance", "eventTime",
		"ProcessWID", "average_response_time")
	details.SetTimeFilter(obj.filter.getData())
	details.SetOrder("eventTime", "ProcessWID")

	data.dataByTime = make(map[time.Time]*dataSourceRow)

	columnByID := make(map[string]int)
	for details.Next(&eventTime,
		&processID, &average_response_time) {

		eventTTime, err := time.ParseInLocation("2006-01-02 15:04:05", eventTime[:19], time.Local)
		checkErr(err)

		if _, exists := columnByID[processID]; !exists {
			columnByID[processID] = len(columnByID)
		}
		if _, exists := data.dataByTime[eventTTime]; !exists {
			data.dataByTime[eventTTime] = &dataSourceRow{cells: make(map[int]float64)}
		}

		d := data.dataByTime[eventTTime]
		d.cells[columnByID[processID]] = 10000 / average_response_time
	}

	data.columns = make([]string, 1+len(columnByID))
	data.columns[0] = `{"id":"","label":"date","type":"datetime"}`
	for name, id := range columnByID {
		data.columns[id+1] = fmt.Sprintf(`{"id":"","label":"%s","type":"number"}`, name)
	}

	return
}

func (obj *WebReporter) getProcessPerformanceStatistics(processId string) (data dataSource) {

	var processID string
	var MIN_cpu, AVG_cpu, MAX_cpu float64
	var MIN_queue_length, AVG_queue_length, MAX_queue_length float64
	var MIN_queue_lengthByCpu, AVG_queue_lengthByCpu, MAX_queue_lengthByCpu float64
	var MIN_memory_performance, AVG_memory_performance, MAX_memory_performance float64
	var MIN_disk_performance, AVG_disk_performance, MAX_disk_performance float64
	var MIN_response_time, AVG_response_time, MAX_response_time float64
	var MIN_average_response_time, AVG_average_response_time, MAX_average_response_time float64

	data.columns = make([]string, 5)
	data.columns[0] = `{"id":"","label":"Process","type":"string"}`
	data.columns[1] = `{"id":"","label":"Minimun","type":"number"}`
	data.columns[2] = `{"id":"","label":"Maximum","type":"number"}`
	data.columns[3] = `{"id":"","label":"Average","type":"number"}`
	data.columns[4] = `{"id":"","label":"Show","type":"boolean"}`

	details := obj.storage.SelectQuery("processesPerformance", "processWID",
		"MIN(cpu)", "AVG(cpu)", "MAX(cpu)",
		"MIN(queue_length)", "AVG(queue_length)", "MAX(queue_length)",
		"MIN(queue_lengthByCpu)", "AVG(queue_lengthByCpu)", "MAX(queue_lengthByCpu)",
		"MIN(memory_performance)", "AVG(memory_performance)", "MAX(memory_performance)",
		"MIN(disk_performance)", "AVG(disk_performance)", "MAX(disk_performance)",
		"MIN(response_time)", "AVG(response_time)", "MAX(response_time)",
		"MIN(average_response_time)", "AVG(average_response_time)", "MAX(average_response_time)",
	)
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processWID = ?", processId)
	details.SetGroup("processWID")
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

	data.rows = make([]string, 7)
	data.rows[0] = fmt.Sprintf(`{"c":[{"v":"cpu"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_cpu, AVG_cpu, MAX_cpu)
	data.rows[1] = fmt.Sprintf(`{"c":[{"v":"queue_length"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_queue_length, AVG_queue_length, MAX_queue_length)
	data.rows[2] = fmt.Sprintf(`{"c":[{"v":"queue_lengthByCpu"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_queue_lengthByCpu, AVG_queue_lengthByCpu, MAX_queue_lengthByCpu)
	data.rows[3] = fmt.Sprintf(`{"c":[{"v":"memory_performance"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_memory_performance, AVG_memory_performance, MAX_memory_performance)
	data.rows[4] = fmt.Sprintf(`{"c":[{"v":"disk_performance"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_disk_performance, AVG_disk_performance, MAX_disk_performance)
	data.rows[5] = fmt.Sprintf(`{"c":[{"v":"response_time"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_response_time, AVG_response_time, MAX_response_time)
	data.rows[6] = fmt.Sprintf(`{"c":[{"v":"average_response_time"},{"v":"%g"},{"v":"%g"},{"v":"%g"},{"v":true}]}`, MIN_average_response_time, AVG_average_response_time, MAX_average_response_time)

	return
}

func (obj *WebReporter) getProcessPerformance(processId string) (data dataSource) {

	var eventTime string
	var cpu, queue_length, queue_lengthByCpu float64
	var memory_performance, disk_performance float64
	var response_time, average_response_time float64

	data.columns = make([]string, 8)
	data.columns[0] = `{"id":"","label":"date","type":"datetime"}`
	data.columns[1] = `{"id":"","label":"cpu","type":"number"}`
	data.columns[2] = `{"id":"","label":"queue_length","type":"number"}`
	data.columns[3] = `{"id":"","label":"queue_lengthByCpu","type":"number"}`
	data.columns[4] = `{"id":"","label":"memory_performance","type":"number"}`
	data.columns[5] = `{"id":"","label":"disk_performance","type":"number"}`
	data.columns[6] = `{"id":"","label":"response_time","type":"number"}`
	data.columns[7] = `{"id":"","label":"average_response_time","type":"number"}`

	data.dataByTime = make(map[time.Time]*dataSourceRow)

	details := obj.storage.SelectQuery("processesPerformance", "eventTime",
		"cpu", "queue_length", "queue_lengthByCpu",
		"memory_performance", "disk_performance",
		"response_time", "average_response_time")
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("processWID = ?", processId)
	for details.Next(&eventTime,
		&cpu, &queue_length, &queue_lengthByCpu,
		&memory_performance, &disk_performance,
		&response_time, &average_response_time) {

		eventTTime, err := time.ParseInLocation("2006-01-02 15:04:05", eventTime[:19], time.Local)
		checkErr(err)

		d := dataSourceRow{cells: make(map[int]float64)}
		d.cells[0] = cpu
		d.cells[1] = queue_length
		d.cells[2] = queue_lengthByCpu
		d.cells[3] = memory_performance
		d.cells[4] = disk_performance
		d.cells[5] = response_time
		d.cells[6] = average_response_time

		data.dataByTime[eventTTime] = &d
	}

	return
}
