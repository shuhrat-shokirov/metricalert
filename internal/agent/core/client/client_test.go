package client

import "testing"

func Test_handler_SendMetric(t *testing.T) {
	type fields struct {
		addr string
	}
	type args struct {
		metricName string
		metricType string
		value      any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test 1",
			fields: fields{
				addr: "http://localhost:8080",
			},
			args: args{
				metricName: "metric1",
				metricType: "gauge",
				value:      1.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &handler{
				addr: tt.fields.addr,
			}
			if err := c.SendMetric(tt.args.metricName, tt.args.metricType, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
