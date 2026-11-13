package model

import (
	"strings"
	"time"
)

const (
	TimezoneWIB  = "WIB"
	TimezoneWITA = "WITA"
	TimezoneWIT  = "WIT"
)

var timezoneByAirport = map[string]string{
	"CGK": TimezoneWIB,
	"SUB": TimezoneWIB,
	"SOC": TimezoneWIB,
	"JOG": TimezoneWIB,
	"BDO": TimezoneWIB,
	"SRG": TimezoneWIB,
	"PDG": TimezoneWIB,
	"PKU": TimezoneWIB,
	"PLM": TimezoneWIB,

	"DPS": TimezoneWITA,
	"UPG": TimezoneWITA,
	"BPN": TimezoneWITA,
	"LOP": TimezoneWITA,
	"MDC": TimezoneWITA,
	"KDI": TimezoneWITA,

	"DJJ": TimezoneWIT,
	"TIM": TimezoneWIT,
	"SOQ": TimezoneWIT,
	"AMQ": TimezoneWIT,
	"LBJ": TimezoneWIT,
}

func AirportTimezone(code string) string {
	return timezoneByAirport[strings.ToUpper(strings.TrimSpace(code))]
}

func FormatAirportDateTime(datetime, airport string) (string, string) {
	timezone := AirportTimezone(airport)
	if timezone == "" {
		return datetime, ""
	}

	value, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return datetime, timezone
	}

	return value.In(timezoneLocation(timezone)).Format(time.RFC3339), timezone
}

func timezoneLocation(timezone string) *time.Location {
	switch timezone {
	case TimezoneWIB:
		return time.FixedZone(TimezoneWIB, 7*60*60)
	case TimezoneWITA:
		return time.FixedZone(TimezoneWITA, 8*60*60)
	case TimezoneWIT:
		return time.FixedZone(TimezoneWIT, 9*60*60)
	default:
		return time.UTC
	}
}
