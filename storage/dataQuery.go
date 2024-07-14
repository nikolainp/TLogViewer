package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type queryResult struct {
	table string
	query string
	args  []any

	isExecuted bool

	data *Storage
	rows *sql.Rows
}

func (obj *Storage) SelectQuery(table string, columns string) interface {
	SetFilter(struct {
		From time.Time
		To   time.Time
	})
	Next(args ...any) bool
} {
	result := new(queryResult)

	result.data = obj
	result.table = table
	result.query = obj.metadata.SelectColumnsSQL(table, columns)
	result.args = make([]any, 0)

	return result
}

// /////////////////////////////////////////////////////////////////////////////
func (obj *queryResult) SetFilter(filter struct{ From, To time.Time }) {
	obj.query = obj.query + " WHERE " + obj.data.metadata.GetFilterSQL(obj.table)
	obj.args = append(obj.args, filter.From)
	obj.args = append(obj.args, filter.To)
}

func (obj *queryResult) Next(args ...any) (ok bool) {
	var err error

	if !obj.isExecuted {
		obj.rows, err = obj.data.db.Query(obj.query, obj.args...)
		if err != nil {
			panic(fmt.Errorf("\nquery: %s\nerror: %w", obj.query, err))
		}
		obj.isExecuted = true
	}

	ok = obj.rows.Next()
	if ok {
		if err = obj.rows.Scan(args...); err != nil {
			panic(fmt.Errorf("\nerror: %w", err))
		}
	}

	return
}
