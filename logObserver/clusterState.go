package logobserver

import (
	"path/filepath"
	"time"
)

type clusterState struct {
	processes  map[string]*clusterProcess
	curProcess *clusterProcess

	storage Storage
}

func (obj *clusterState) init(storage Storage) {
	obj.processes = make(map[string]*clusterProcess)

	obj.storage = storage
}

func (obj *clusterState) addEvent(data event) {

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

func (obj *clusterState) FlushAll() {
	for _, process := range obj.processes {
		obj.storage.WriteProcess(
			process.name,
			process.catalog,
			process.process,
			0,
			0,
			process.firstEventTime,
			process.lastEventTime,
		)
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

	obj.firstEventTime = data.eventStopTime
	obj.lastEventTime = data.eventStopTime

	return obj
}

func (obj *clusterProcess) addEvent(data event) {
	if obj.firstEventTime.After(data.eventStopTime) {
		obj.firstEventTime = data.eventStopTime
	}
	if obj.lastEventTime.Before(data.eventStopTime) {
		obj.lastEventTime = data.eventStopTime
	}
}
