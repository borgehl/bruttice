package util

import (
	"fmt"
	"time"
)

var TZLocation *time.Location = time.UTC

var cstTZ *time.Location

func init() {
	var err error
	cstTZ, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
}

func SetTZLocation(loc *time.Location) {
	TZLocation = loc
}

func NowDate() time.Time {
	now := time.Now().In(TZLocation)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, TZLocation).UTC()
}

func ParseDateTime(datetimeStr string, formatString string, today time.Time) (time.Time, error) {
	ret, err := time.Parse(formatString, datetimeStr)
	if err != nil {
		return ret, err
	}
	if ret.Unix() < today.Unix() {
		ret = ret.AddDate(1, 0, 0)
	}
	return ret, nil
}

func MultipleDatetimeParsingTries(datetimeStr string, formats []string, today time.Time) (ret time.Time, err error) {
	for _, format := range formats {
		ret, err = ParseDateTime(datetimeStr, format, today)
		if err == nil {
			return
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse %s from formats %+v", datetimeStr, formats)
}

func DisplayDateFormat(dt time.Time) string {
	period := ""
	if dt.Format("Jan") != dt.Format("January") {
		period = "."
	}
	return dt.Format(fmt.Sprintf("Monday Jan%s 2, 2006", period))
}

func ReplaceTimezone(t time.Time, loc *time.Location) time.Time {
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		loc,
	)
}

func DatetimeIsCentral(t time.Time) time.Time {
	return ReplaceTimezone(t, cstTZ)
}
