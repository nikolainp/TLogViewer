package filewalker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type CancelFunc func() bool
type Monitor interface {
	WriteEvent(frmt string, args ...any)
	NewData(size int64)
	FinishedData(count, size int64)
}
type EventWalkFunc func(string, string, string)

type pathWalker struct {
	rootPath string
	check    lineChecker

	isCancel  CancelFunc
	monitor   Monitor
	eventWalk EventWalkFunc
}

///////////////////////////////////////////////////////////////////////////////

func GetFileWalker(isCancelFunc CancelFunc, monitor Monitor) (obj pathWalker) {
	obj.isCancel = isCancelFunc
	obj.monitor = monitor

	obj.check.init(isCancelFunc, monitor)

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
	obj.monitor.WriteEvent("Data catalog: %s\n", basePath)

	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			obj.monitor.WriteEvent("Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		obj.monitor.NewData(info.Size())
		obj.processFile(path)

		if obj.isCancel() {
			return fmt.Errorf("process is cancel")
		}

		return nil
	})
	if err != nil {
		obj.monitor.WriteEvent("Error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) processFile(fileName string) {
	var subFileName string
	var err error

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

///////////////////////////////////////////////////////////////////////////////
