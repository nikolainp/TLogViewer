package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type queryResult struct {
	table                    string
	query                    string
	queryWhere, queryGroupBy []string
	args                     []any

	isExecuted bool

	data *Storage
	rows *sql.Rows
}

func (obj *Storage) SelectQuery(table string, columns ...string) interface {
	SetTimeFilter(struct {
		From time.Time
		To   time.Time
	})
	SetFilter(filter ...string)
	SetGroup(fields ...string)
	Next(args ...any) bool
} {
	result := new(queryResult)

	result.data = obj
	result.table = table
	result.query = obj.metadata.SelectColumnsSQL(table, columns...)
	result.args = make([]any, 0)

	return result
}

// /////////////////////////////////////////////////////////////////////////////
func (obj *queryResult) SetTimeFilter(filter struct{ From, To time.Time }) {
	obj.queryWhere = append(obj.queryWhere, obj.data.metadata.GetFilterSQL(obj.table))
	obj.args = append(obj.args, filter.From)
	obj.args = append(obj.args, filter.To)
}
func (obj *queryResult) SetFilter(filter ...string) {
	if len(filter) != 0 {
		obj.queryWhere = append(obj.queryWhere, filter[0])
	}
	if len(filter) == 2 {
		obj.args = append(obj.args, filter[1])
	}
}

func (obj *queryResult) SetGroup(fields ...string) {
	obj.queryGroupBy = append(obj.queryGroupBy, strings.Join(fields, ", "))
}

func (obj *queryResult) Next(args ...any) (ok bool) {
	var err error

	makeQuery := func() string {
		query := obj.query
		if len(obj.queryWhere) != 0 {
			query = query + " WHERE " + strings.Join(obj.queryWhere, " AND ")
		}
		if len(obj.queryGroupBy) != 0 {
			query = query + " GROUP BY " + strings.Join(obj.queryGroupBy, ", ")
		}

		return query
	}

	if !obj.isExecuted {
		query := makeQuery()
		obj.rows, err = obj.data.db.Query(query, obj.args...)
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
