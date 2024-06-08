package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Monitor interface {
	WriteEvent(frmt string, args ...any)
	NewData(count int, size int64)
	ProcessedData(count int, size int64)
	IsCancel() bool
	Cancel() chan bool
}

type QueryResult interface {
	Next(args ...any) bool
}

type Storage struct {
	metadata metaData

	cacheInsertValueSQL map[string]string

	monitor Monitor
	db      *sql.DB
}

type queryResult struct {
	data *sql.Rows
}

// Конструктор Storage
func CreateCache(monitor Monitor) (*Storage, error) {

	var err error

	obj := newStorage(monitor)
	if obj.db, err = openDB(""); err != nil {
		return nil, err
	}

	initDB(obj.db, obj.metadata, true)

	return obj, nil
}

func Open(stroragePath string) (obj *Storage, err error) {

	if _, err := os.Stat(stroragePath); err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}

	obj = newStorage(nil)
	if obj.db, err = openDB(stroragePath); err != nil {
		return nil, err
	}

	return obj, nil
}

func (obj *Storage) FlushAll(stroragePath string) error {

	if err := os.Remove(stroragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear storage: %v", err)
	}

	db, err := openDB(stroragePath)
	if err != nil {
		return err
	}
	initDB(db, obj.metadata, false)
	db.Close()

	obj.saveAll(stroragePath)

	return nil

}

func (obj *Storage) SelectAll(table string, columns string) interface{ Next(args ...any) bool } {
	query := obj.metadata.SelectColumnsSQL(table, columns)
	rows, err := obj.db.Query(query)
	if err != nil {
		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
	}
	return &queryResult{data: rows}
}

func (obj *Storage) WriteRow(table string, args ...any) {
	query, ok := obj.cacheInsertValueSQL[table]
	if !ok {
		query = obj.metadata.GetInsertValueSQL(table)
		obj.cacheInsertValueSQL[table] = query
	}

	if _, err := obj.db.Exec(query, args...); err != nil {
		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
	}
}

func (obj *Storage) Update(table string, args ...any) {

	fields := make([]any, 0, len(args))
	values := make([]any, 0, len(args))

	for i := range args {
		if i%2 == 0 {
			fields = append(fields, args[i])
		} else {
			values = append(values, args[i])
		}
	}

	query := obj.metadata.GetUpdateSQL(table, fields...)

	if _, err := obj.db.Exec(query, values...); err != nil {

		qq := "PRAGMA table_list"
		rows, err := obj.db.Query(qq)
		if err != nil {
			panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
		}

		for rows.Next() {
			var schema, name, ttype string
			var ncol int
			var wr, strict bool
			rows.Scan(&schema, &name, &ttype, &ncol, &wr, &strict)

			fmt.Printf("name: %s\n", name)
		}

		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
	}
}

// func (obj *Storage) SetIdByGroup(table string, column, group string) {
// 	query := obj.metadata.SetIdByGroup(table, column, group)

// 	if _, err := obj.db.Exec(query); err != nil {
// 		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
// 	}

// }

///////////////////////////////////////////////////////////////////////////////

func (obj *queryResult) Next(args ...any) (ok bool) {

	ok = obj.data.Next()
	if ok {
		if err := obj.data.Scan(args...); err != nil {
			panic(fmt.Errorf("\nerror: %w", err))
		}
	}

	return
}

///////////////////////////////////////////////////////////////////////////////

func newStorage(monitor Monitor) *Storage {
	obj := new(Storage)
	obj.metadata = NewMetadata()
	obj.cacheInsertValueSQL = make(map[string]string)
	obj.monitor = monitor

	return obj
}

func openDB(stroragePath string) (*sql.DB, error) {
	var dataSource string
	var err error

	if stroragePath == "" {
		dataSource = ":memory:?mode=memory&cache=private&nolock=1&psow=1"
	} else {
		dataSource = "file:" + stroragePath + "?cache=private&nolock=1&psow=1"
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

func initDB(db *sql.DB, meta metaData, isCache bool) {
	for _, table := range meta.InitDB(isCache) {
		if _, err := db.Exec(table); err != nil {
			panic(err)
		}
	}
}

func (obj *Storage) saveAll(dbPath string) {
	if _, err := obj.db.Exec("ATTACH DATABASE '" + dbPath + "' AS datafile"); err != nil {
		panic(err)
	}

	parts := obj.metadata.SaveAll("datafile")
	obj.monitor.NewData(len(parts), 0)
	for _, part := range parts {
		if _, err := obj.db.Exec(part); err != nil {
			panic(fmt.Errorf("query: %s\nerror: %w", part, err))
		}
		obj.monitor.ProcessedData(1, 0)
	}

	if _, err := obj.db.Exec("DETACH datafile"); err != nil {
		panic(err)
	}
}
