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

	//obj.storage.SetIdByGroup("processesPerformance", "processID", "process, pid")

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

///////////////////////////////////////////////////////////////////////////////

func getSimpleProperty(data string, name string) string {
	start := strings.Index(data, name)
	if start == -1 {
		return ""
	}
	start += len(name)

	length := strings.Index(data[start:], ",")
	if length == -1 {
		return data[start:]
	}

	return data[start : start+length]
}

func getSubString(data string, start string, finish string) string {
	var startPos, finishPos int

	if startPos = strings.Index(data, start); startPos == -1 {
		startPos = 0
	} else {
		startPos += len(start)
	}

	finishPos = strings.Index(data[startPos:], finish)
	if finishPos == -1 {
		return data[startPos:]
	}

	return data[startPos : startPos+finishPos]
}

func isIPAddress(data string) bool {
	isNumber := func(data byte) bool {
		if data == '0' || data == '1' || data == '2' || data == '3' ||
			data == '4' || data == '5' || data == '6' || data == '7' ||
			data == '8' || data == '9' {
			return true
		}

		return false
	}

	if len(data) == 0 {
		return false
	}
	if data[0] == '[' && data[len(data)-1] == ']' {
		// IPv6
		return true
	}

	maybePoint := false
	for i := range data {
		if !maybePoint && isNumber(data[i]) {
			maybePoint = true
			continue
		}
		if maybePoint && data[i] == '.' {
			maybePoint = false
			continue
		}
		if !isNumber(data[i]) {
			return false
		}
	}

	return true
}
