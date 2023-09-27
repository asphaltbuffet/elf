package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddCmd(t *testing.T) {
	got := GetAddCmd()

	checkCommand(t, got, "add")
}

func Test_validYearCompletionArgs(t *testing.T) {
	type args struct {
		args       []string
		toComplete string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "default",
			args: args{
				args:       []string{},
				toComplete: "",
			},
			want: []string{"2015", "2016", "2017", "2018", "2019", "2020", "2021", "2022"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := validYearCompletionArgs(nil, tt.args.args, tt.args.toComplete)
			assert.Equal(t, tt.want, got)
		})
	}
}
