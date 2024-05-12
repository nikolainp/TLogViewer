package logobserver

import "time"

type event struct {
	catalog   string
	fileName  string
	eventData string

	eventTime time.Time
}
type chanEvents chan event

type processor struct {
	clusterState clusterState

	storage Storage

	isCancel CancelFunc
}

func (obj *processor) init(isCancelFunc CancelFunc, storage Storage) {
	obj.isCancel = isCancelFunc

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

			data.addProperties()
			obj.clusterState.addEvent(data)
		}

		if obj.isCancel() {
			return
		}
	}
}

func (obj *event) addProperties() {
	obj.eventTime = time.Now()
}
