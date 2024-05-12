package filewalker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type Monitor interface {
	WriteEvent(frmt string, args ...any)
	NewData(size int64)
	FinishedData(count, size int64)
	IsCancel() bool
}
type EventWalkFunc func(string, string, string)

type pathWalker struct {
	rootPath string
	check    lineChecker

	filePaths chan string

	monitor   Monitor
	eventWalk EventWalkFunc
}

///////////////////////////////////////////////////////////////////////////////

func New(monitor Monitor) (obj pathWalker) {
	obj.monitor = monitor

	obj.check.init(monitor)

	return
}

///////////////////////////////////////////////////////////////////////////////

func (obj *pathWalker) Walk(basePath string, eventWalk EventWalkFunc) {
	obj.rootPath = basePath
	obj.eventWalk = eventWalk

	obj.startWalk(basePath)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *pathWalker) startWalk(basePath string) {
	var wg sync.WaitGroup

	obj.monitor.WriteEvent("Data catalog: %s\n", basePath)

	goFunc := func(work func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			work()
		}()
	}

	obj.filePaths = make(chan string, 1000)
	goFunc(func() { obj.processFile() })

	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			obj.monitor.WriteEvent("Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		obj.monitor.NewData(info.Size())
		obj.filePaths <- path
		//obj.processFile(path)

		if obj.monitor.IsCancel() {
			return fmt.Errorf("process is cancel")
		}

		return nil
	})
	if err != nil {
		obj.monitor.WriteEvent("Error walking the path %q: %v\n", basePath, err)
	}

	close(obj.filePaths)
	wg.Wait()
}

func (obj *pathWalker) processFile() {
	var subFileName string
	var err error

	doWork := func(fileName string) {
		subFileName, err = filepath.Rel(obj.rootPath, fileName)
		if err != nil {
			subFileName = fileName
		}

		fileStream, err := os.Open(fileName)
		if err != nil {
			obj.monitor.WriteEvent("Error open: %q: %v\n", fileName, err)
		}
		defer fileStream.Close()
		obj.check.processStream(subFileName, fileStream, obj.eventWalk)
	}

	for {
		select {
		case fileName, ok := <-obj.filePaths:
			if !ok {
				return
			}
			doWork(fileName)
		}

		if obj.monitor.IsCancel() {
			return
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
