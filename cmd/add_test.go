package cmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetAddCmd(t *testing.T) {
	got := GetAddCmd()
	assert.NotNil(t, got)
	assert.IsType(t, &cobra.Command{}, got)
	assert.Equal(t, "add", got.Name())
}

func Test_runAdd(t *testing.T) {
	type args struct {
		args           []string
		day            int
		implementation string
	}

	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dayArg = tt.args.day
			langArg = tt.args.implementation

			err := runAdd(tt.args.args)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func Test_validateInput(t *testing.T) {
	type args struct {
		args           []string
		day            int
		implementation string
	}

	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "good",
			args: args{
				args:           []string{"2015"},
				day:            1,
				implementation: "go",
			},
			assertion: assert.NoError,
		},
		{
			name: "year is not int",
			args: args{
				args:           []string{"fake"},
				day:            1,
				implementation: "go",
			},
			assertion: assert.Error,
			errText:   "invalid year",
		},
		{
			name: "year is too low",
			args: args{
				args:           []string{"42"},
				day:            1,
				implementation: "go",
			},
			assertion: assert.Error,
			errText:   "year is out of range",
		},
		{
			name: "year is too high",
			args: args{
				args:           []string{"5000"},
				day:            1,
				implementation: "go",
			},
			assertion: assert.Error,
			errText:   "year is out of range: 5000 >",
		},
		{
			name: "day is too low",
			args: args{
				args:           []string{"2015"},
				day:            0,
				implementation: "go",
			},
			assertion: assert.Error,
			errText:   "day is out of range",
		},
		{
			name: "day is too high",
			args: args{
				args:           []string{"2015"},
				day:            26,
				implementation: "go",
			},
			assertion: assert.Error,
			errText:   "day is out of range",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dayArg = tt.args.day
			langArg = tt.args.implementation

			err := validateAddInput(nil, tt.args.args)
			tt.assertion(t, err)

			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
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

func Test_getValidYears(t *testing.T) {
	got := getValidYears()
	assert.GreaterOrEqual(t, len(got), 7)
}

func Test_getMaxYear(t *testing.T) {
	assert.LessOrEqual(t, getMaxYear(), time.Now().Year())
}
