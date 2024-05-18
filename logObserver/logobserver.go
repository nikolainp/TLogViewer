package logobserver

import (
	"sync"
	"time"
)

type Monitor interface {
	WriteEvent(frmt string, args ...any)
	NewData(size int64)
	FinishedData(count, size int64)
	IsCancel() bool
	Cancel() chan bool
}

type Storage interface {
	WriteDetails(title string) error
	WriteProcess(name, catalog, process string, pid, port int, firstEvent, lastEvent time.Time) error
}

type supervisor struct {
	worker *processor
	events chanEvents

	storage Storage

	monitor Monitor
	wg      sync.WaitGroup
}

func New(monitor Monitor, storage Storage, title string) (obj *supervisor) {

	obj = new(supervisor)
	obj.events = make(chanEvents)
	obj.storage = storage
	obj.monitor = monitor

	goFunc := func(work func()) {
		obj.wg.Add(1)
		go func() {
			defer obj.wg.Done()
			work()
		}()
	}

	obj.worker = new(processor)
	obj.worker.init(obj.monitor, storage, title)
	goFunc(func() { obj.worker.start(obj.events) })

	return obj
}

func (obj *supervisor) FlushAll() {
	close(obj.events)
	obj.wg.Wait()

	obj.worker.FlushAll()
}

func (obj *supervisor) ConsiderEvent(catalog string, fileName string, eventData string) {
	obj.events <- event{catalog: catalog, fileName: fileName, eventData: eventData}
}
