package logobserver

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type clusterState struct {
	processes  map[string]*clusterProcess
	curProcess *clusterProcess
	// processID  map[string]int

	storage Storage
}

func (obj *clusterState) init(storage Storage) {
	obj.processes = make(map[string]*clusterProcess)
	// obj.processID = make(map[string]int)

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
		obj.curProcess.processID = len(obj.processes)
	}
}

func (obj *clusterState) flushAll() error {

	{
		getServerName := func(data map[string]bool) string {
			for ip := range data {
				for _, process := range obj.processes {
					if len(process.server) == 0 {
						continue
					}
					if _, ok := process.ip[ip]; ok {
						return process.server
					}
				}
			}
			return ""
		}

		for _, process := range obj.processes {
			if len(process.server) > 0 {
				continue
			}
			process.server = getServerName(process.ip)
		}
	}

	obj.agentStandardCallFinish()

	mapToString := func(data map[string]bool) string {
		res := make([]string, 0, len(data))
		for key := range data {
			res = append(res, key)
		}
		return strings.Join(res, "; ")
	}

	for _, process := range obj.processes {
		obj.storage.WriteRow("processes",
			process.name, process.catalog, process.process,
			process.processID,
			process.processType,
			process.pid, process.port,
			process.UID,
			process.server, mapToString(process.ip),
			process.firstEventTime, process.lastEventTime,
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

	obj.storage.WriteRow("processesPerformance", data.stopTime,
		process, pid, cpu, queue_length, queue_lengthByCpu,
		memory_performance, disk_performance,
		response_time, average_response_time)
}

func (obj *clusterState) agentStandardCallFinish() {

	rows := obj.storage.SelectQuery("processesPerformance", "process, pid")
	for {
		var serverName, pid string

		ok := rows.Next(&serverName, &pid)
		if !ok {
			break
		}
		serverName = strings.Replace(serverName, "tcp://", "", 1)
		port := getSimpleProperty(serverName, ":")
		serverName = strings.Replace(serverName, ":"+port, "", 1)

		for _, process := range obj.processes {
			if strings.Compare(process.server, serverName) == 0 &&
				process.pid == pid {
				process.port = port
				break
			}
		}

	}
	for _, process := range obj.processes {
		obj.storage.Update("processesPerformance", "processID", process.processID,
			"process", "tcp://"+process.server+":"+process.port, "pid", process.pid)
	}

}

///////////////////////////////////////////////////////////////////////////////

type clusterProcess struct {
	name, catalog, process string

	processID int

	processType string
	pid, port   string
	UID         string
	server      string
	ip          map[string]bool

	contexts map[string]*serverContext

	firstEventTime time.Time
	lastEventTime  time.Time
}

func newClusterProcess(data event) *clusterProcess {
	obj := new(clusterProcess)
	obj.name = data.catalog
	obj.catalog = filepath.Dir(data.catalog)
	obj.process = filepath.Base(data.catalog)

	obj.processType = getSimpleProperty(data.eventData, ",process=")
	obj.pid = getSimpleProperty(obj.process, "_")

	obj.ip = make(map[string]bool)

	obj.contexts = make(map[string]*serverContext)

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

	// server name
	{
		isTrueEvent := func(data event) bool {
			return data.eventType == "CONN" &&
				strings.Contains(data.eventData, "'addrBelongsToThisComputer2,") &&
				strings.Contains(data.eventData, ", result=true'")
		}

		if isTrueEvent(data) {
			address := getSimpleProperty(data.eventData, ", address=")

			if isIPAddress(address) {
				if strings.Compare(address, "[::1]") != 0 &&
					strings.Compare(address, "127.0.0.1") != 0 {
					obj.ip[address] = true
				}
			} else {
				if len(obj.server) == 0 {
					obj.server = address
				}
			}
		}
	}
	// server ip
	{
		isTrueEvent := func(data event) bool {
			return data.eventType == "CONN" &&
				strings.Contains(data.eventData, ",Protected=0,")
		}

		if isTrueEvent(data) {
			address := getSimpleProperty(data.eventData, "Connected, client=")
			address = getSubString(address, ")", ":")
			if isIPAddress(address) {
				if strings.Compare(address, "[::1]") != 0 &&
					strings.Compare(address, "127.0.0.1") != 0 {
					obj.ip[address] = true
				}
			}
		}
	}
	// server context
	{
		isTrueEvent := func(data event) bool {
			return data.eventType == "SCOM"
		}

		// 45:01.730000-0,SCOM,2,process=rphost,OSThread=1960,t:clientID=115,t:applicationName=JobScheduler,Func='new ServerProcessData(20a67c5cb40,,)'
		// 45:01.730002-0,SCOM,3,process=rphost,OSThread=1960,t:clientID=115,t:applicationName=JobScheduler,Func='setSrcProcessName(20a67c5cb40,IBName,IBName)'
		// 50:02.234010-300504009,SCOM,2,process=rphost,OSThread=1960,t:clientID=115,t:applicationName=JobScheduler,ProcessName=IBName,SrcProcessName=IBName,Func='delete ServerProcessData(20a67c5cb40,IBName,IBName)'

		// 19:00.478000-0,SCOM,2,process=rphost,OSThread=2292,t:clientID=55,Func='new ServerProcessData(222e2279ec0,,)'
		// 19:00.478002-0,SCOM,3,process=rphost,OSThread=2292,t:clientID=55,Func='setSrcProcessName(222e2279ec0,IBName,IBName)'
		// 19:05.135006-0,SCOM,3,process=rphost,p:processName=IBName,p:processName=IBName2d006f19-1b79-4dfa-a087-c888c028c120,OSThread=2292,t:clientID=55,t:applicationName=BackgroundJob,t:computerName=computerName,t:connectID=1,Func='changeServerContextName(222e2279ec0,IBName,IBName2d006f19-1b79-4dfa-a087-c888c028c120)->success'
		// 19:05.260001-4782000,SCOM,2,process=rphost,OSThread=2292,t:clientID=55,ProcessName=IBName,SrcProcessName=IBName,ProcessName=IBName2d006f19-1b79-4dfa-a087-c888c028c120,Func='delete ServerProcessData(222e2279ec0,IBName,IBName2d006f19-1b79-4dfa-a087-c888c028c120)'

		if isTrueEvent(data) {
			funcValue := getComplexProperty(data.eventData, ",Func=")
			name := getSubString(funcValue, "'", "(")
			value := strings.Split(getSubString(funcValue, "(", ")"), ",")

			switch name {
			case "new ServerProcessData":
			case "setSrcProcessName":
				obj.contexts[value[0]] = &serverContext{
					id: value[0], name: value[1],
					createTime: data.stopTime,
				}
			case "changeServerContextName":
				if context, ok := obj.contexts[value[0]]; ok {
					context.renameTime = data.stopTime
				} else {
					obj.contexts[value[0]] = &serverContext{
						id: value[0], name: value[1],
						renameTime: data.stopTime,
					}
				}
			case "delete ServerProcessData":
				if context, ok := obj.contexts[value[0]]; ok {
					context.renameTime = data.stopTime
				} else {
					obj.contexts[value[0]] = &serverContext{
						id: value[0], name: value[1],
						createTime: data.startTime,
						deleteTime: data.stopTime,
					}
				}
			default:
				fmt.Printf("func = %s", name)
			}
			// if isIPAddress(address) {
			// 	if strings.Compare(address, "[::1]") != 0 &&
			// 		strings.Compare(address, "127.0.0.1") != 0 {
			// 		obj.ip[address] = true
			// 	}
			// }
		}

	}
}

///////////////////////////////////////////////////////////////////////////////

type serverContext struct {
	id, name                           string
	createTime, renameTime, deleteTime time.Time
}
