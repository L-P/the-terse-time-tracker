package util_test

import (
	"testing"
	"time"
	"tt/internal/util"
)

var startOfDayTests = []struct {
	v, expected string
}{
	{"2021-04-18T10:16:42+02:00", "2021-04-18T00:00:00+02:00"},
	{"2021-04-18T01:02:02+02:00", "2021-04-18T00:00:00+02:00"},
	{"2021-04-18T23:59:59+02:00", "2021-04-18T00:00:00+02:00"},
	{"2021-04-18T00:00:00+02:00", "2021-04-18T00:00:00+02:00"},
	{"2021-04-18T10:16:42Z", "2021-04-18T00:00:00Z"},
}

func TestGetStartOfDay(t *testing.T) {
	for _, test := range startOfDayTests {
		in, err := time.Parse(time.RFC3339, test.v)
		if err != nil {
			t.Fatal(err)
		}

		if actual := util.GetStartOfDay(in).Format(time.RFC3339); actual != test.expected {
			t.Errorf("\n%s\n%s", actual, test.expected)
		}
	}
}

var startOfWeekTests = []struct {
	v, expected string
}{
	{"2021-04-11T00:00:00+02:00", "2021-04-05T00:00:00+02:00"},
	{"2021-04-12T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-13T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-14T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-15T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-16T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-17T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-18T00:00:00+02:00", "2021-04-12T00:00:00+02:00"},
	{"2021-04-19T00:00:00+02:00", "2021-04-19T00:00:00+02:00"},
}

func TestGetStartOfWeek(t *testing.T) {
	for _, test := range startOfWeekTests {
		in, err := time.Parse(time.RFC3339, test.v)
		if err != nil {
			t.Fatal(err)
		}

		if actual := util.GetStartOfWeek(in).Format(time.RFC3339); actual != test.expected {
			t.Errorf("\n%s\n%s", actual, test.expected)
		}
	}
}
