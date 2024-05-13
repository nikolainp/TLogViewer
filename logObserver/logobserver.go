package logobserver

import (
	"sync"
)

type Monitor interface {
	WriteEvent(frmt string, args ...any)
	NewData(size int64)
	FinishedData(count, size int64)
	IsCancel() bool
}

type Storage interface {
	WriteProcess(name, catalog, process string, pid, port int) error
}

type supervisor struct {
	worker *processor
	events chanEvents

	storage Storage

	monitor Monitor
	wg      sync.WaitGroup
}

func New(monitor Monitor, storage Storage) (obj *supervisor) {

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
	obj.worker.init(obj.monitor, storage)
	goFunc(func() { obj.worker.start(obj.events) })

	return obj
}

func (obj *supervisor) FlushAll() {
	close(obj.events)
	obj.wg.Wait()
}

func (obj *supervisor) ConsiderEvent(catalog string, fileName string, eventData string) {
	obj.events <- event{catalog: catalog, fileName: fileName, eventData: eventData}
}
