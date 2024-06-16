package webreporter

import (
	"strings"
	"text/template"
	"time"
)

type dataFilter struct {
	minimumTime, startTime  time.Time
	maximumTime, finishTime time.Time
}

func (obj *dataFilter) setTime(start, finish time.Time) {
	obj.minimumTime = start
	obj.startTime = start
	obj.maximumTime = finish
	obj.finishTime = finish
}

func (obj *dataFilter) getContent() string {
	w := new(strings.Builder)
	dataFilter, err := template.New("dataFilter").Parse(dataFilterTemplate)
	checkErr(err)

	data := struct {
		MinimumTime, StartTime  string
		MaximumTime, FinishTime string
	}{
		MinimumTime: obj.minimumTime.Format("2006-01-02T15:04"),
		StartTime:   obj.startTime.Format("2006-01-02T15:04"),
		MaximumTime: obj.maximumTime.Format("2006-01-02T15:04"),
		FinishTime:  obj.finishTime.Format("2006-01-02T15:04"),
	}

	err = dataFilter.Execute(w, data)
	checkErr(err)

	return w.String()
}

const dataFilterTemplate = `
<form method="post">
  <fieldset>
  	<!-- <legend>Do you agree to the terms?</legend> -->
  	<label>Данные отобраны с:<input type="datetime-local" name="TimeFrom" 
		min="{{.MinimumTime}}" max="{{.MaximumTime}}"
		value="{{.StartTime}}"
		autocomplete="name" /></label>
  	<label>по:<input type="datetime-local" name="TimeTo" 
		min="{{.MinimumTime}}" max="{{.MaximumTime}}"
		value="{{.FinishTime}}"
		autocomplete="name" /></label>
	<button>Ок</button>
  </fieldset>
</form>
`
