package webreporter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type dataSource struct {
	columns    []string
	rows       []string
	dataByTime map[time.Time]*dataSourceRow
}

type dataSourceRow struct {
	cells map[int]float64
}

func (obj *WebReporter) dataSource(w http.ResponseWriter, req *http.Request) {
	var js string

	section := req.PathValue("section")
	source := req.PathValue("source")

	js = `
	{
	"cols": [
			%s
		],
	"rows": [
			%s
		]
 	}
`

	toDataRows := func(data dataSource) (result []string) {

		maxValues := 1000
		columns := len(data.columns)
		maxRows := 1 + maxValues/columns

		result = make([]string, 0, maxRows)

		keys := make([]time.Time, 0, len(data.dataByTime))
		for k := range data.dataByTime {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Before(keys[j])
		})

		beginTime := keys[0]
		finishTime := keys[len(keys)-1]
		duration := time.Duration(finishTime.Sub(beginTime).Seconds()/float64(maxRows)) * time.Second

		beginTime = beginTime.Add(duration)
		dataRow := make(map[int]float64)
		dataCount := make(map[int]int)
		for i := range keys {

			if keys[i].Before(beginTime) {
				for j, k := range data.dataByTime[keys[i]].cells {
					dataRow[j] = dataRow[j] + k
					dataCount[j] = dataCount[j] + 1
				}
			} else {
				beginTime = beginTime.Add(duration)

				
				for  j := 0; j < len(data.columns); j++{
					if d, exists := dataRow[j]; exists {
						
					} else {

					}
				}
				result = append(result, "")

				dataRow = make(map[int]float64)
				dataCount = make(map[int]int)
			}
		}

		return
	}

	switch section {
	case "root":
		data := obj.getProcesses()
		js = fmt.Sprintf(js, strings.Join(data.columns, ","), strings.Join(data.rows, ","))
	case "performance":
		processID := req.Header.Get("ID")
		switch source {
		case "statistics.json":
			data := obj.getPerformanceStatistics()
			js = fmt.Sprintf(js, strings.Join(data.columns, ","), strings.Join(data.rows, ","))
		case "data.json":
			js = processID
			data := obj.getPerformance()
			js = fmt.Sprintf(js, strings.Join(data.columns, ","),
				strings.Join(toDataRows(data), ","))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(js))
}
