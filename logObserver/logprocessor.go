package logobserver

import (
	"fmt"
	"time"
)

type event struct {
	catalog   string
	fileName  string
	eventData string

	eventStopTime time.Time
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
			if obj.firstEventTime.After(data.eventStopTime) || obj.firstEventTime.IsZero() {
				obj.firstEventTime = data.eventStopTime
			}
			if obj.lastEventTime.Before(data.eventStopTime) || obj.lastEventTime.IsZero() {
				obj.lastEventTime = data.eventStopTime
			}

			obj.clusterState.addEvent(data)
		}
	}
}

func (obj *processor) FlushAll() {

	if err := obj.storage.WriteDetails(obj.title, obj.version,
		obj.eventDataSize, 1000 * obj.eventDataSize/time.Since(obj.startProcessingTime).Milliseconds(),
		time.Now(), obj.firstEventTime, obj.lastEventTime); err != nil {
		obj.monitor.WriteEvent("error: %w\n", err)
		return
	}

	if err := obj.clusterState.FlushAll(); err != nil {
		obj.monitor.WriteEvent("error: %w\n", err)
		return
	}

}

func (obj *event) addProperties() error {
	if len(obj.eventData) < 12 {
		return fmt.Errorf("short event")
	}
	strLineTime := obj.fileName + ":" + obj.eventData[:12]

	stopTime, err := time.ParseInLocation("06010215.log:04:05", string(strLineTime), time.Local)
	if err != nil {
		return err
	}

	obj.eventStopTime = stopTime

	return nil
}
