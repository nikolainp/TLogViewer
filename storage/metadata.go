package storage

import (
	"database/sql"
	"fmt"
	"strings"
)

type metaData struct {
	tables []metaTable
}

type metaTable struct {
	name    string
	columns []metaColumn
	//insertStm *sql.Stmt
	isCache bool
}

type metaColumn struct {
	name      string
	datatype  string
	isCache   bool
	isService bool
}

///////////////////////////////////////////////////////////////////////////////

func (obj *metaData) init() {
	obj.tables = []metaTable{
		{name: "details",
			columns: []metaColumn{
				{name: "title", datatype: "TEXT"}, {name: "version", datatype: "TEXT"},
				{name: "processingSize", datatype: "NUMBER"},
				{name: "processingSpeed", datatype: "NUMBER"}, {name: "processingTime", datatype: "NUMBER"},
				{name: "firstEventTime", datatype: "DATETIME"}, {name: "lastEventTime", datatype: "DATETIME"},
			},
		},
		{name: "processes",
			columns: []metaColumn{
				{name: "name", datatype: "TEXT"}, {name: "catalog", datatype: "TEXT"}, {name: "process", datatype: "TEXT"},
				{name: "pid", datatype: "NUMBER"}, {name: "port", datatype: "NUMBER"},
				{name: "firstEventTime", datatype: "DATETIME"}, {name: "lastEventTime", datatype: "DATETIME"},
				{name: "processID", datatype: "NUMBER"}, {name: "server", datatype: "TEXT"},
			},
		},
		{name: "processesPerfomance",
			columns: []metaColumn{
				{name: "processID", datatype: "NUMBER", isService: true}, {name: "eventTime", datatype: "DATATIME"},
				{name: "process", datatype: "TEXT", isCache: true}, {name: "pid", datatype: "TEXT", isCache: true},
				{name: "cpu", datatype: "NUMBER"},
				{name: "queue_length", datatype: "NUMBER"},
				{name: "queue_lengthByCpu", datatype: "NUMBER"},
				{name: "memory_performance", datatype: "NUMBER"},
				{name: "disk_performance", datatype: "NUMBER"},
				{name: "response_time", datatype: "NUMBER"},
				{name: "average_response_time", datatype: "REAL"},
			},
		},
	}

}

func (obj *metaData) initDB(db *sql.DB, isCache bool) {

	for i := range obj.tables {
		if obj.tables[i].isCache && !isCache {
			continue
		}
		if _, err := db.Exec(obj.tables[i].getCreateSQL(true)); err != nil {
			panic(err)
		}
	}
}

func (obj *metaData) saveAll(db *sql.DB, dbPath string) {

	if _, err := db.Exec("ATTACH DATABASE '" + dbPath + "' AS datafile"); err != nil {
		panic(err)
	}

	for i := range obj.tables {
		if obj.tables[i].isCache {
			continue
		}
		query := obj.tables[i].getInsertSelectSQL()
		if _, err := db.Exec(query); err != nil {
			panic(fmt.Errorf("query: %s\nerror: %w", query, err))
		}
	}

	{
		var seq, name, file string
		query :=
			`PRAGMA database_list`

		rows, err := db.Query(query)
		if err != nil {
			err = fmt.Errorf("SelectDetails: %w", err)
			return
		}

		for rows.Next() {
			rows.Scan(&seq, &name, &file)
			fmt.Printf("DB: %s %s %s\n", seq, name, file)
		}
	}

	if _, err := db.Exec("DETACH datafile"); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func (obj *metaTable) getCreateSQL(isCache bool) string {

	queryColumns := make([]string, len(obj.columns))
	for i := range obj.columns {
		if obj.columns[i].isCache && !isCache {
			continue
		}
		queryColumns[i] = fmt.Sprintf("%s %s", obj.columns[i].name, obj.columns[i].datatype)
	}

	return fmt.Sprintf("CREATE TABLE %s (%s)", obj.name, strings.Join(queryColumns, ","))
}

func (obj *metaTable) getInsertValueSQL() string {
	queryColumns := make([]string, 0, len(obj.columns))
	queryValues := make([]string, 0, len(obj.columns))

	for i := range obj.columns {
		if obj.columns[i].isService {
			continue
		}
		queryColumns = append(queryColumns, obj.columns[i].name)
		queryValues = append(queryValues, "?")
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", obj.name,
		strings.Join(queryColumns, ","),
		strings.Join(queryValues, ","),
	)
}

func (obj *metaTable) getInsertSelectSQL() string {
	queryColumns := make([]string, 0, len(obj.columns))

	for i := range obj.columns {
		if obj.columns[i].isCache {
			continue
		}
		queryColumns = append(queryColumns, obj.columns[i].name)
	}

	return fmt.Sprintf("INSERT INTO datafile.%s (%s) SELECT %s FROM %s",
		obj.name, strings.Join(queryColumns, ","),
		strings.Join(queryColumns, ","), obj.name,
	)
}
