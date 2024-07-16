package webreporter

import (
	"net/http"
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

func (obj *dataFilter) getContent(url string) string {
	w := new(strings.Builder)
	dataFilter, err := template.New("dataFilter").Parse(dataFilterTemplate)
	checkErr(err)

	data := struct {
		Url                     string
		MinimumTime, StartTime  string
		MaximumTime, FinishTime string
	}{
		Url:         url,
		MinimumTime: obj.minimumTime.Format("2006-01-02T15:04"),
		StartTime:   obj.startTime.Format("2006-01-02T15:04"),
		MaximumTime: obj.maximumTime.Format("2006-01-02T15:04"),
		FinishTime:  obj.finishTime.Format("2006-01-02T15:04"),
	}

	err = dataFilter.Execute(w, data)
	checkErr(err)

	return w.String()
}

func (obj *dataFilter) setContext(w http.ResponseWriter, req *http.Request) {

	url := req.PostFormValue("url")
	obj.startTime, _ = time.ParseInLocation("2006-01-02T15:04", req.PostFormValue("TimeFrom"), time.Local)
	obj.finishTime, _ = time.ParseInLocation("2006-01-02T15:04", req.PostFormValue("TimeTo"), time.Local)

	http.Redirect(w, req, url, http.StatusSeeOther)
}

func (obj *dataFilter) getData() (filter struct{ From, To time.Time }) {
	filter.From = obj.startTime
	filter.To = obj.finishTime

	return
}

func (obj *dataFilter) getStartTime(tt time.Time) time.Time {
	if obj.startTime.Before(tt) {
		return tt
	}

	return obj.startTime
}

func (obj *dataFilter) getFinishTime(tt time.Time) time.Time {
	if obj.finishTime.After(tt) {
		return tt
	}

	return obj.finishTime
}

const dataFilterTemplate = `
<form method="post" action="/datafilter">
  <fieldset>
	<input type="hidden" name="url" value="{{.Url}}" />
  	<label for="TimeFrom">Данные отобраны с:</label>
	<input type="datetime-local" id="TimeFrom" name="TimeFrom" 
		min="{{.MinimumTime}}" max="{{.MaximumTime}}"
		value="{{.StartTime}}" style="width:150px"/>
  	<label for="TimeTo">по:</label>
	<input type="datetime-local" id="TimeTo" name="TimeTo" 
		min="{{.MinimumTime}}" max="{{.MaximumTime}}"
		value="{{.FinishTime}}" style="width:150px"/>
	<button type="submit">Ок</button>
	<button type="reset">Cancel</button>
  </fieldset>
</form>
`
