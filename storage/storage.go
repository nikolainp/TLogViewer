package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Конструктор Storage
func New(storagePath string) (*Storage, error) {

	dbPath := filepath.Clean(storagePath) + ".db"
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("clear storage: %v", err)
	}

	db := new(Storage)
	if err := db.create(dbPath); err != nil {
		return nil, err
	}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func Open(storagePath string) (*Storage, error) {

	dbPath := filepath.Clean(storagePath) + ".db"
	db := new(Storage)
	if err := db.create(dbPath); err != nil {
		return nil, err
	}

	return db, nil
}

type Process struct {
	Name    string
	Catalog string
	Process string
	Pid     int
	Port    int
}

func (obj *Storage) SelectAllProcesses() (data []Process, err error) {
	query := "SELECT name, catalog, process, pid, port FROM processes"

	var name string
	var catalog string
	var process string
	var pid int
	var port int

	data = make([]Process, 0)

	rows, err := obj.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("selectAllProcess: %w", err)
	}

	for rows.Next() {
		rows.Scan(&name, &catalog, &process, &pid, &port)
		data = append(data, Process{name, catalog, process, pid, port})
	}

	return
}

func (obj *Storage) WriteProcess(name, catalog, process string, pid, port int) error {
	query := "INSERT INTO processes (name, catalog, process, pid, port) VALUES (?,?,?,?,?)"

	if _, err := obj.db.Exec(query, name, catalog, process, pid, port); err != nil {
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
		"CREATE TABLE processes (name TEXT, catalog text, process TEXT, pid NUMBER, port NUMER)",
	}

	for i := range queries {
		_, err := obj.db.Exec(queries[i])
		return err
	}

	return nil
}
