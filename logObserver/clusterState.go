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
		if err := obj.storage.WriteProcess(
			process.name,
			process.catalog,
			process.process,
			0,
			0,
			process.firstEventTime,
			process.lastEventTime,
		); err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *clusterState) agentStandardCall(data event) {

	var process, pid string

	isTrueEvent := func(data event) bool {
		return data.eventType == "CLSTR" &&
			strings.Contains(data.eventData, ",process=rmngr,") &&
			strings.Contains(data.eventData, ",Event=Performance update,")
	}
	saveCounters := func(processID int, eventTime time.Time, dataFields []string) error {
		for _, field := range dataFields {
			subFields := strings.Split(field, "=")
			if subFields[0] == "process" || subFields[0] == "pid" {
				continue
			}

			if value, err := strconv.ParseFloat(subFields[1], 32); err != nil {
				err = obj.storage.WriteProcessPerfomance(processID, eventTime, subFields[0], value)
				if err != nil {
					return err
				}
			}

		}
		return nil
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

	}

	processName := process + "\n" + pid
	if processID, ok := obj.processID[processName]; ok {
		saveCounters(processID, data.stopTime, dataFields)
	} else {
		processID = len(obj.processID) + 1
		obj.processID[processName] = processID
		saveCounters(processID, data.stopTime, dataFields)
	}
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
