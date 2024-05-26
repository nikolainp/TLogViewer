package logobserver

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type clusterState struct {
	processes  map[string]*clusterProcess
	curProcess *clusterProcess
	processID  map[string]int

	storage Storage
}

func (obj *clusterState) init(storage Storage) {
	obj.processes = make(map[string]*clusterProcess)
	obj.processID = make(map[string]int)

	obj.storage = storage
}

func (obj *clusterState) addEvent(data event) {

	obj.agentStandardCall(data)

	if obj.curProcess != nil {
		if obj.curProcess.name == data.catalog {
			obj.curProcess.addEvent(data)
			return
		}
	}

	if process, ok := obj.processes[data.catalog]; ok {
		process.addEvent(data)
	} else {
		obj.curProcess = newClusterProcess(data)
		obj.processes[data.catalog] = obj.curProcess
	}
}

func (obj *clusterState) flushAll() error {
	for _, process := range obj.processes {
		obj.storage.WriteRow("processes",
			process.name,
			process.catalog,
			process.process,
			0,
			0,
			process.firstEventTime, process.lastEventTime,
			0, "",
		)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *clusterState) agentStandardCall(data event) {

	var process, pid string
	var cpu, queue_length, queue_lengthByCpu int
	var memory_performance, disk_performance int
	var response_time int
	var average_response_time float64

	isTrueEvent := func(data event) bool {
		return data.eventType == "CLSTR" &&
			strings.Contains(data.eventData, ",process=rmngr,") &&
			strings.Contains(data.eventData, ",Event=Performance update,")
	}
	parseInt := func(data string) int {
		val, _ := strconv.Atoi(data)
		return val
	}
	parseFloat := func(data string) float64 {
		val, _ := strconv.ParseFloat(data, 32)
		return val
	}

	if !isTrueEvent(data) {
		return
	}

	dataPosition := strings.Index(data.eventData, ",Data='")
	if dataPosition == -1 {
		return
	}

	dataStr := data.eventData[dataPosition+7:]
	dataFields := strings.Split(dataStr, ",")
	for _, field := range dataFields {
		subFields := strings.Split(field, "=")
		if subFields[0] == "process" {
			process = subFields[1]
		}
		if subFields[0] == "pid" {
			pid = subFields[1]
		}
		if subFields[0] == "cpu" {
			cpu = parseInt(subFields[1])
		}
		if subFields[0] == "queue_length" {
			queue_length = parseInt(subFields[1])
		}
		if subFields[0] == "queue_length/cpu_num" {
			queue_lengthByCpu = parseInt(subFields[1])
		}
		if subFields[0] == "memory_performance" {
			memory_performance = parseInt(subFields[1])
		}
		if subFields[0] == "disk_performance" {
			disk_performance = parseInt(subFields[1])
		}
		if subFields[0] == "response_time" {
			response_time = parseInt(subFields[1])
		}
		if subFields[0] == "average_response_time" {
			average_response_time = parseFloat(subFields[1])
		}
	}

	obj.storage.WriteRow("processesPerfomance", data.stopTime,
		process, pid, cpu, queue_length, queue_lengthByCpu,
		memory_performance, disk_performance,
		response_time, average_response_time)
}

///////////////////////////////////////////////////////////////////////////////

type clusterProcess struct {
	name, catalog, process string
	//	pid, port              int
	firstEventTime time.Time
	lastEventTime  time.Time
}

func newClusterProcess(data event) *clusterProcess {
	obj := new(clusterProcess)
	obj.name = data.catalog
	obj.catalog = filepath.Dir(data.catalog)
	obj.process = filepath.Base(data.catalog)

	obj.firstEventTime = data.stopTime
	obj.lastEventTime = data.stopTime

	return obj
}

func (obj *clusterProcess) addEvent(data event) {
	if obj.firstEventTime.After(data.stopTime) {
		obj.firstEventTime = data.stopTime
	}
	if obj.lastEventTime.Before(data.stopTime) {
		obj.lastEventTime = data.stopTime
	}
}
