package logobserver

import (
	"reflect"
	"testing"
)

func Test_clusterProcess_addEvent(t *testing.T) {
	tests := []struct {
		name string
		args event
		obj  clusterProcess
	}{
		{
			"test 1",
			event{catalog: "srv1\\proc_1", fileName: "24041408.log", eventData: "32:47.733006-0,EXCPCNTX,"},
			clusterProcess{name: "srv1\\proc_1", catalog: "srv1", process: "proc_1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newClusterProcess(tt.args)
			if !reflect.DeepEqual(*got, tt.obj) {
				t.Errorf("clusterProcess() = %v, want %v", got, tt.obj)
			}
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("clusterProcess() error = %T, wantErr %T", err, tt.wantErr)
			// }
		})
	}
}
