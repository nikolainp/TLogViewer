package config

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantObj Config
		wantErr bool
	}{
		{"test 1", []string{"programname"}, Config{
			programName: "programname",
			DataPath:    "",
		}, true},
		{
			"test 2",
			[]string{"programname", "a1"},
			Config{
				programName: "programname",
				DataPath:    "a1",
			},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotObj, err := New(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotObj, tt.wantObj) {
				t.Errorf("GetConfig() = %v, want %v", gotObj, tt.wantObj)
			}
		})
	}
}
