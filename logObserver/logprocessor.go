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

	isCancel CancelFunc
}

func (obj *processor) init(isCancelFunc CancelFunc) {
	obj.isCancel = isCancelFunc

	obj.clusterState.init()
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
