package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_UpdateGauge1(t *testing.T) {

	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "first test",
			args: args{
				name:  "test",
				value: 1.1,
			},
			want: 1.1,
		},
		{
			name: "second test",
			args: args{
				name:  "test",
				value: 0.9,
			},
			want: 0.9,
		},
		{
			name: "third test",
			args: args{
				name:  "test",
				value: -1,
			},
			want: -1,
		},
	}

	s := NewMemStorage(&Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.UpdateGauge(tt.args.name, tt.args.value)
			assert.Equal(t, tt.want, s.gauges[tt.args.name])
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "first test",
			args: args{
				name:  "test",
				value: 1,
			},
			want: 1,
		},
		{
			name: "second test",
			args: args{
				name:  "test",
				value: 1000012,
			},
			want: 1000013,
		},
		{
			name: "third test",
			args: args{
				name:  "test",
				value: -1,
			},
			want: 1000012,
		},
		{
			name: "invalid test",
			args: args{
				name:  "test",
				value: 10001,
			},
			wantErr: true,
			want:    1000012,
		},
	}

	s := NewMemStorage(&Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.UpdateCounter(tt.args.name, tt.args.value)
			if tt.wantErr {
				assert.NotEqual(t, tt.want, s.counters[tt.args.name])
				return
			}

			assert.Equal(t, tt.want, s.counters[tt.args.name])
		})
	}
}
