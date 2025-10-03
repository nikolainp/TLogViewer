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
		columns := len(data.columns) - 1
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
		dataRow := make([]float64, columns)
		dataCount := make([]float64, columns)
		dataStr := make([]string, columns)
		for i := range keys {

			if keys[i].Before(beginTime) {
				for j, k := range data.dataByTime[keys[i]].cells {
					dataRow[j] = dataRow[j] + k
					dataCount[j] = dataCount[j] + 1
				}
			} else {
				for i := 0; i < len(dataRow); i++ {
					if dataCount[i] == 0 {
						dataStr[i] = `null`
					} else {
						dataStr[i] = fmt.Sprintf("%.2f", dataRow[i]/dataCount[i])
					}
					dataRow[i] = 0
					dataCount[i] = 0
				}
				result = append(result, fmt.Sprintf(
					`{"c":[{"v":"Date(%s)"},{"v":%s}]}`,
					beginTime.Format("2006, 01, 02, 15, 04, 05"),
					strings.Join(dataStr, `},{"v":`),
				))
				beginTime = beginTime.Add(duration)
			}
		}

		return
	}

	switch section {
	case "root":
		data := obj.getProcesses()
		js = fmt.Sprintf(js, strings.Join(data.columns, ","), strings.Join(data.rows, ","))
	case "processes":
		data := obj.getProcessesLiveTime()
		js = fmt.Sprintf(js, strings.Join(data.columns, ","), strings.Join(data.rows, ","))
	case "performance":
		processID := req.Header.Get("ID")
		switch source {
		case "statistics.json":
			data := obj.getPerformanceStatistics(processID)
			js = fmt.Sprintf(js, strings.Join(data.columns, ","), strings.Join(data.rows, ","))
		case "data.json":
			data := obj.getPerformance(processID)
			js = fmt.Sprintf(js, strings.Join(data.columns, ","),
				strings.Join(toDataRows(data), ","))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(js))
}
