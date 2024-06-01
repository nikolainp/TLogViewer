package logobserver

import (
	"reflect"
	"testing"
	"time"
)

func Test_event_addProperties(t *testing.T) {
	tests := []struct {
		name    string
		obj     event
		want    event
		wantErr bool
	}{
		{"test 1", event{}, event{}, true},
		{"test 2", event{
			catalog:   "rphost_1",
			fileName:  "24052110.log",
			eventData: "17:49.627001-3000000,CLSTR,0,process=rmngr,p:processName=RegMngrCntxt,",
		}, event{
			catalog:   "rphost_1",
			fileName:  "24052110.log",
			eventData: "0,process=rmngr,p:processName=RegMngrCntxt,",
			startTime: time.Date(2024, 5, 21, 10, 17, 46, 627001000, time.Local),
			stopTime:  time.Date(2024, 5, 21, 10, 17, 49, 627001000, time.Local),
			eventType: "CLSTR",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.want.duration = tt.want.stopTime.Sub(tt.want.startTime)

			if err := tt.obj.addProperties(); (err != nil) != tt.wantErr {
				t.Errorf("event.addProperties() error = %v, wantErr  %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.obj, tt.want) {
				t.Errorf("get: \n%v\n, want: \n%v", tt.obj, tt.want)
			}

		})
	}
}

func Test_isIPAddress(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "test 1", arg: "", want: false},
		{name: "test 2", arg: "aaaa", want: false},
		{name: "test 3", arg: "aaaa.loc", want: false},
		{name: "test 4", arg: "[::1]", want: true},
		{name: "test 5", arg: "10.0.0.1", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIPAddress(tt.arg); got != tt.want {
				t.Errorf("isIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
