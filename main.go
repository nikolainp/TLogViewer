package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolainp/TLogViewer/config"
	filewalker "github.com/nikolainp/TLogViewer/fileWalker"
	logobserver "github.com/nikolainp/TLogViewer/logObserver"
	"github.com/nikolainp/TLogViewer/monitor"
	"github.com/nikolainp/TLogViewer/storage"
	webreporter "github.com/nikolainp/TLogViewer/webReporter"
)

var (
	version = "dev"
	//	commit  = "none"
	date = "unknown"
)

var cancelChan chan bool

func init() {
	signChan := make(chan os.Signal, 10)
	cancelChan = make(chan bool, 1)

	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		signal := <-signChan
		// Run Cleanup
		fmt.Fprintf(os.Stderr, "\nCaptured %v, stopping and exiting...\n", signal)
		cancelAndExit()
	}()
}

func main() {
	var storage *storage.Storage

	conf := getConfig(os.Args)

	if !conf.ShowReportOnly {
		monitor := monitor.New(isCancel)
		walker := filewalker.New(isCancel, monitor)

		monitor.Start()
		storage = getNewStorage(conf.DataPath)
		observer := logobserver.New(isCancel, storage)
		walker.Walk(conf.DataPath, observer.ConsiderEvent)
		observer.FlushAll()
		monitor.Stop()
	} else {
		storage = getOldStorage(conf.DataPath)
	}

	reporter := webreporter.New(storage)
	reporter.Start()
}

///////////////////////////////////////////////////////////////////////////////

func getConfig(args []string) config.Config {
	conf, err := config.New(args)

	if err != nil {
		switch err := err.(type) {
		case config.PrintVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case config.PrintUsage:
			fmt.Fprint(os.Stderr, err.Usage)
		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		os.Exit(0)
	}

	return conf
}

func cancelAndExit() {
	cancelChan <- true
	close(cancelChan)
	os.Exit(0)
}

func isCancel() bool {
	select {
	case _, ok := <-cancelChan:
		return !ok
	default:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////

func getNewStorage(path string) *storage.Storage {

	db, err := storage.New(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Storage error: %v\n", err)
		cancelAndExit()
	}

	return db
}

func getOldStorage(path string) *storage.Storage {

	db, err := storage.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Storage error: %v\n", err)
		cancelAndExit()
	}

	return db
}
