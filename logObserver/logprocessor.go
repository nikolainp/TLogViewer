package logobserver

import (
	"fmt"
	"strings"
	"time"
)

type event struct {
	catalog   string
	fileName  string
	eventData string

	startTime time.Time
	stopTime  time.Time
	duration  time.Duration
	eventType string
}
type chanEvents chan event

type processor struct {
	clusterState clusterState

	storage Storage
	title   string
	version string

	startProcessingTime time.Time
	eventDataSize       int64
	firstEventTime      time.Time
	lastEventTime       time.Time

	monitor Monitor
}

func (obj *processor) init(monitor Monitor, storage Storage, title, version string) {
	obj.monitor = monitor
	obj.storage = storage
	obj.title = title
	obj.version = version

	obj.startProcessingTime = time.Now()

	obj.clusterState.init(obj.storage)
}

func (obj *processor) start(events chanEvents) {
	for {
		select {
		case _, ok := <-obj.monitor.Cancel():
			if !ok {
				return
			}
		case data, ok := <-events:
			if !ok {
				return
			}

			if err := data.addProperties(); err != nil {
				obj.monitor.WriteEvent("error event: %s: %s:\n%s\n%w\n", data.catalog, data.fileName, data.eventData, err)
				continue
			}

			obj.eventDataSize += int64(len(data.eventData))
			if obj.firstEventTime.After(data.stopTime) || obj.firstEventTime.IsZero() {
				obj.firstEventTime = data.stopTime
			}
			if obj.lastEventTime.Before(data.stopTime) || obj.lastEventTime.IsZero() {
				obj.lastEventTime = data.stopTime
			}

			obj.clusterState.addEvent(data)
		}
	}
}

func (obj *processor) FlushAll() {
	obj.storage.WriteRow("details", obj.title, obj.version,
		obj.eventDataSize, 1000*obj.eventDataSize/time.Since(obj.startProcessingTime).Milliseconds(),
		time.Now(), obj.firstEventTime, obj.lastEventTime)

	if err := obj.clusterState.flushAll(); err != nil {
		obj.monitor.WriteEvent("error: %w\n", err)
		return
	}

	//obj.storage.SetIdByGroup("processesPerfomance", "processID", "process, pid")

	{

		type processID struct {
			id int
			process, pid string
		}
		data := make([]processID, 0)

		rows := obj.storage.SelectAll("processesPerfomance", "process, pid")
		for {
			var row processID

			row.id = len(data) + 1
			ok := rows.Next(&row.process, &row.pid)
			if !ok {
				break
			}
			data = append(data, row)
		}
		for _, row := range data {
			obj.storage.Update("processesPerfomance", "processID", row.id, "process", row.process, "pid", row.pid)
		}
	}
}

func (obj *event) addProperties() (err error) {
	if len(obj.eventData) < 12 {
		return fmt.Errorf("short event")
	}
	strLineTime := obj.fileName + ":" + obj.eventData[:12]

	obj.stopTime, err = time.ParseInLocation("06010215.log:04:05", string(strLineTime), time.Local)
	if err != nil {
		return err
	}

	commaPosition := strings.Index(obj.eventData[13:], ",")
	if commaPosition == -1 {
		return fmt.Errorf("error format event: " + obj.eventData)
	}
	obj.duration, err = time.ParseDuration(obj.eventData[13:13+commaPosition] + "us")
	if err != nil {
		return err
	}

	obj.startTime = obj.stopTime.Add(-1 * obj.duration)

	dataStartPosition := 13 + commaPosition + 1
	commaPosition = strings.Index(obj.eventData[dataStartPosition:], ",")
	if commaPosition == -1 {
		return fmt.Errorf("error format event: " + obj.eventData)
	}
	obj.eventType = obj.eventData[dataStartPosition : dataStartPosition+commaPosition]

	dataStartPosition = dataStartPosition + commaPosition + 1
	obj.eventData = obj.eventData[dataStartPosition:]

	return nil
}
