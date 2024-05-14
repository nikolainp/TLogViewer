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

	monitor Monitor
}

func (obj *processor) init(monitor Monitor, storage Storage) {
	obj.monitor = monitor

	obj.storage = storage

	obj.clusterState.init(obj.storage)
}

func (obj *processor) start(events chanEvents) {
	for {
		select {
		case data, ok := <-events:
			if !ok {
				return
			}

			if err := data.addProperties(); err != nil {
				obj.monitor.WriteEvent("error event: %s: %s:\n%s\n%w\n", data.catalog, data.fileName, data.eventData, err)
				continue
			}
			obj.clusterState.addEvent(data)
		}

		if obj.monitor.IsCancel() {
			return
		}
	}
}

func (obj *processor) FlushAll() {

	obj.clusterState.FlushAll()

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
