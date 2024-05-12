package logobserver

import (
	"sync"
)

type Storage interface {
	WriteProcess(name, catalog, process string, pid, port int) error
}

type CancelFunc func() bool

type supervisor struct {
	worker *processor
	events chanEvents

	storage Storage

	isCancel CancelFunc
	wg       sync.WaitGroup
}

func New(isCancelFunc CancelFunc, storage Storage) (obj *supervisor) {

	obj = new(supervisor)
	obj.events = make(chanEvents)
	obj.storage = storage
	obj.isCancel = isCancelFunc

	goFunc := func(work func()) {
		obj.wg.Add(1)
		go func() {
			defer obj.wg.Done()
			work()
		}()
	}

	obj.worker = new(processor)
	obj.worker.init(obj.isCancel)
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
