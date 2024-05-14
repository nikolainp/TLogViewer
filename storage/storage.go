package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Конструктор Storage
func New(stroragePath string) (*Storage, error) {

	if err := os.Remove(stroragePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("clear storage: %v", err)
	}

	db := new(Storage)
	if err := db.create(stroragePath); err != nil {
		return nil, err
	}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func Open(stroragePath string) (*Storage, error) {

	if _, err := os.Stat(stroragePath); err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}

	db := new(Storage)
	if err := db.create(stroragePath); err != nil {
		return nil, err
	}

	return db, nil
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

func (obj *Storage) SelectAllProcesses() (data []Process, err error) {
	query := "SELECT name, catalog, process, pid, port, firstEventTime, lastEventTime FROM processes"

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

///////////////////////////////////////////////////////////////////////////////

func (obj *Storage) create(stroragePath string) error {
	var err error

	obj.db, err = sql.Open("sqlite3", stroragePath)
	if err != nil {
		return fmt.Errorf("open storage: %v", err)
	}
	err = obj.db.Ping()
	if err != nil {
		return fmt.Errorf("ping storage: %v", err)
	}

	return nil
}

func (obj *Storage) init() error {
	queries := []string{
		`CREATE TABLE processes (
			name TEXT, catalog text, process TEXT, 
			pid NUMBER, port NUMER, 
			firstEventTime DATETIME, lastEventTime DATETIME)`,
	}

	for i := range queries {
		_, err := obj.db.Exec(queries[i])
		return err
	}

	return nil
}
