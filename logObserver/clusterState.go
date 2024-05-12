package logobserver

import "path/filepath"

type clusterState struct {
	processes map[string]*clusterProcess
	curProcess *clusterProcess
}

func (obj *clusterState) init() {
	obj.processes = make(map[string]*clusterProcess)
}

func (obj *clusterState) addEvent(data event) {

	if obj.curProcess.name == data.catalog {
		return
	}

	if process, ok := obj.processes[data.catalog]; ok {
		process.addEvent(data)
	} else {
		obj.curProcess = newClusterProcess(data)
		obj.processes[data.catalog] = obj.curProcess
	}
}

///////////////////////////////////////////////////////////////////////////////

type clusterProcess struct {
	name, catalog, process string
//	pid, port              int
}

func newClusterProcess(data event) *clusterProcess {
	obj := new(clusterProcess)
	obj.name = data.catalog
	obj.catalog = filepath.Dir(data.catalog)
	obj.process = filepath.Base(data.catalog)

	return obj
}

func (obj *clusterProcess) addEvent(data event) {
	
}
