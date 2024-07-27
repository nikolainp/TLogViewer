package storage

import (
	"fmt"
	"strings"
)

type metaData interface {
	InitDB(isCache bool) []string
	SaveAll(schema string) []string

	GetInsertValueSQL(table string) string
	GetUpdateSQL(table string, fields ...any) string
	//	SetIdByGroup(table string, column, group string) string

	SelectColumnsSQL(table string, columns string) string
	GetFilterSQL(table string) string
}

type implMetaData struct {
	tables map[string]metaTable
}

type metaTable struct {
	name    string
	columns []metaColumn

	//insertStm *sql.Stmt
	isCache bool

	columnTimeFrom, columnTimeTo string
}

type metaColumn struct {
	name     string
	datatype string

	isCache   bool
	isService bool

	isTimeFrom bool
	isTimeTo   bool
}

///////////////////////////////////////////////////////////////////////////////

func NewMetadata() metaData {
	obj := new(implMetaData)
	obj.init()
	return obj
}

func (obj *implMetaData) InitDB(isCache bool) []string {
	queries := make([]string, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.isCache && !isCache {
			continue
		}
		queries = append(queries, table.getCreateSQL(isCache))
	}

	return queries
}

func (obj *implMetaData) SaveAll(schema string) []string {
	queries := make([]string, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.isCache {
			continue
		}
		queries = append(queries, table.getInsertSelectSQL())
	}

	return queries
}

func (obj *implMetaData) GetInsertValueSQL(table string) string {
	return obj.tables[table].getInsertValueSQL()
}

func (obj *implMetaData) SelectColumnsSQL(table string, columns string) (query string) {

	if len(columns) == 0 {
		names := make([]string, 0, 10)
		metaColumns := obj.tables[table].columns
		for i := range metaColumns {
			names = append(names, metaColumns[i].name)
		}
		columns = strings.Join(names, ", ")
	}

	query = fmt.Sprintf("SELECT DISTINCT %s FROM %s", columns, table)

	return
}

func (obj *implMetaData) GetFilterSQL(table string) (filter string) {

	metaTable := obj.tables[table]
	return fmt.Sprintf("%s >= ? and %s <= ?", metaTable.columnTimeTo, metaTable.columnTimeFrom)
}

func (obj *implMetaData) GetUpdateSQL(table string, fields ...any) (query string) {

	where := make([]string, 0, len(fields))
	for i := range fields {
		if i == 0 {
			continue
		}
		where = append(where, fmt.Sprintf("%s = ?", fields[i]))
	}

	if len(where) == 0 {
		query = fmt.Sprintf("UPDATE %s SET %s = ? ", table, fields[0])
	} else {
		query = fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s", table, fields[0], strings.Join(where, " AND "))
	}

	return
}

///////////////////////////////////////////////////////////////////////////////

func (obj *implMetaData) init() {
	obj.tables = map[string]metaTable{
		"details": {name: "details",
			columns: []metaColumn{
				{name: "title", datatype: "TEXT"}, {name: "version", datatype: "TEXT"},
				{name: "processingSize", datatype: "NUMBER"},
				{name: "processingSpeed", datatype: "NUMBER"}, {name: "processingTime", datatype: "DATETIME"},
				{name: "firstEventTime", datatype: "DATETIME"}, {name: "lastEventTime", datatype: "DATETIME"},
			},
		},
		"processes": {name: "processes",
			columns: []metaColumn{
				{name: "name", datatype: "TEXT"}, {name: "catalog", datatype: "TEXT"}, {name: "process", datatype: "TEXT"},
				{name: "processID", datatype: "NUMBER"},
				{name: "processType", datatype: "TEXT"},
				{name: "pid", datatype: "TEXT"}, {name: "port", datatype: "TEXT"},
				{name: "UID", datatype: "TEXT"},
				{name: "serverName", datatype: "TEXT"}, {name: "IP", datatype: "TEXT"},
				{name: "firstEventTime", datatype: "DATETIME", isTimeFrom: true},
				{name: "lastEventTime", datatype: "DATETIME", isTimeTo: true},
			},
		},
		"processesPerformance": {name: "processesPerformance",
			columns: []metaColumn{
				{name: "processID", datatype: "NUMBER", isService: true},
				{name: "eventTime", datatype: "DATATIME", isTimeFrom: true, isTimeTo: true},
				{name: "process", datatype: "TEXT", isCache: false}, {name: "pid", datatype: "TEXT", isCache: false},
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

	for tableName, table := range obj.tables {
		for _, column := range table.columns {
			if column.isTimeFrom {
				table.columnTimeFrom = column.name
			}
			if column.isTimeTo {
				table.columnTimeTo = column.name
			}
		}
		obj.tables[tableName] = table
	}

	return

}

func (obj *metaTable) getCreateSQL(isCache bool) string {

	queryColumns := make([]string, 0, len(obj.columns))
	for i := range obj.columns {
		if obj.columns[i].isCache && !isCache {
			continue
		}
		queryColumns = append(queryColumns, fmt.Sprintf("%s %s", obj.columns[i].name, obj.columns[i].datatype))
	}

	return fmt.Sprintf("CREATE TABLE %s (%s)", obj.name, strings.Join(queryColumns, ","))
}

func (obj metaTable) getInsertValueSQL() string {
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
