package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	metadata metaData

	db *sql.DB
}

// Конструктор Storage
func CreateCache() (*Storage, error) {

	var err error

	obj := newStorage()
	if obj.db, err = openDB(""); err != nil {
		return nil, err
	}

	obj.metadata.initDB(obj.db, true)

	return obj, nil
}

func Open(stroragePath string) (obj *Storage, err error) {

	if _, err := os.Stat(stroragePath); err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}

	obj = newStorage()
	if obj.db, err = openDB(stroragePath); err != nil {
		return nil, err
	}


	return obj, nil
}

func (obj *Storage) FlushAll(stroragePath string) error {

	if err := os.Remove(stroragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear storage: %v", err)
	}

	db, err := openDB(stroragePath); 
	if err != nil {
		return err
	}
	obj.metadata.initDB(db, false)
	db.Close()

	obj.metadata.saveAll(obj.db, stroragePath)

	return nil

}

type Process struct {
	Name           string
	Catalog        string
	Process        string
	Pid            int
	Port           int
	FirstEventTime time.Time
	LastEventTime  time.Time
}

func (obj *Storage) SelectTitle() (title string, err error) {
	query :=
		`SELECT title
		FROM details 
		LIMIT 1`

	rows, err := obj.db.Query(query)
	if err != nil {
		err = fmt.Errorf("SelectDetails: %w", err)
		return
	}

	rows.Next()
	rows.Scan(&title)

	return
}

func (obj *Storage) SelectDetails() (title, version string, processingSize, processingSpeed int64, processingTime, firstEventTime, lastEventTime time.Time, err error) {
	query :=
		`SELECT title, version, processingSize, processingSpeed, processingTime, firstEventTime, lastEventTime
		FROM details 
		LIMIT 1`

	rows, err := obj.db.Query(query)
	if err != nil {
		err = fmt.Errorf("SelectDetails: %w", err)
		return
	}

	rows.Next()
	rows.Scan(&title, &version, &processingSize, &processingSpeed, &processingTime, &firstEventTime, &lastEventTime)

	return
}

func (obj *Storage) WriteDetails(title, version string, processingSize, processingSpeed int64, processingTime, firstEventTime, lastEventTime time.Time) error {
	query := `INSERT INTO details 
		(title, version, processingSize, processingSpeed, processingTime, firstEventTime, lastEventTime) 
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	if _, err := obj.db.Exec(query,
		title, version,
		processingSize, processingSpeed,
		processingTime, firstEventTime, lastEventTime); err != nil {

		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))

		return fmt.Errorf("writeProcess: %w", err)
	}

	return nil
}

func (obj *Storage) SelectAllProcesses() (data []Process, err error) {
	query :=
		`SELECT name, catalog, process, pid, port, firstEventTime, lastEventTime 
		FROM processes
		ORDER BY catalog, process`

	var name string
	var catalog string
	var process string
	var pid int
	var port int
	var firstEvent, lastEvent time.Time

	data = make([]Process, 0)

	rows, err := obj.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("selectAllProcess: %w", err)
	}

	for rows.Next() {
		rows.Scan(&name, &catalog, &process, &pid, &port, &firstEvent, &lastEvent)
		data = append(data, Process{name, catalog, process, pid, port, firstEvent, lastEvent})
	}

	return
}

func (obj *Storage) WriteProcess(name, catalog, process string, pid, port int, firstEvent, lastEvent time.Time) error {
	query := "INSERT INTO processes (name, catalog, process, pid, port, firstEventTime, lastEventTime) VALUES (?,?,?,?,?,?,?)"

	if _, err := obj.db.Exec(query,
		name, catalog, process,
		pid, port,
		firstEvent, lastEvent); err != nil {
		return fmt.Errorf("writeProcess: %w", err)
	}

	return nil
}

func (obj *Storage) WriteProcessPerfomance(processID int, eventTime time.Time, counter string, value float64) error {
	query := "INSERT INTO processesPerfomance (processID, eventTime, counterName, counterValue) VALUES (?,?,?,?)"

	if _, err := obj.db.Exec(query,
		processID, eventTime, counter, value); err != nil {
		return fmt.Errorf("WriteProcessPerfomance: %w", err)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func  openDB(stroragePath string) (*sql.DB, error) {
	var dataSource string
	var err error

	if stroragePath == "" {
		dataSource = ":memory:?mode=memory&cache=private&nolock=1&psow=1"
	} else {
		dataSource = "file:"+stroragePath+"?cache=private&nolock=1&psow=1"
	}
	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping storage: %v", err)
	}

	queries := []string{
		`PRAGMA main.journal_mode = MEMORY`,
	}
	for i := range queries {
		if _, err := db.Exec(queries[i]); err != nil {
			panic(err)
		}
	}


	return db, nil
}

func newStorage() *Storage {
	obj := new(Storage)
	obj.metadata.init()

	return obj
}
