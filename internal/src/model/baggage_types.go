package model

import (
	"regexp"
	"strings"
)

var baggageWeightPattern = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(kg)\b`)

func NormalizeBaggage(baggage Baggage) Baggage {
	return Baggage{
		CarryOn: normalizeBaggageValue(baggage.CarryOn),
		Checked: normalizeBaggageValue(baggage.Checked),
	}
}

func normalizeBaggageValue(value string) string {
	value = strings.TrimSpace(value)
	return baggageWeightPattern.ReplaceAllString(value, "$1 kg")
}
