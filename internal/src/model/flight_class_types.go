package model

import "strings"

// cabinMap maps lowercase single-letter IATA booking codes to lowercase cabin names
var cabinMap = map[string]string{
	// first class
	"f": "first",
	"a": "first",
	"p": "first",

	// business class
	"j": "business",
	"c": "business",
	"d": "business",
	"i": "business",
	"z": "business",

	// premium economy
	"w": "premium economy",
	"e": "premium economy",

	// economy class
	"y": "economy",
	"b": "economy",
	"m": "economy",
	"h": "economy",
	"k": "economy",
	"l": "economy",
	"q": "economy",
	"v": "economy",
	"s": "economy",
	"t": "economy",
	"u": "economy",
	"g": "economy",
	"n": "economy",
	"o": "economy",
	"x": "economy",
}

// GetCabinClass converts a booking code into a lowercase cabin name
func GetCabinClass(code string) string {
	cleaned := strings.TrimSpace(strings.ToLower(code))
	if cleaned == "" {
		return "unknown"
	}

	// Extract only the first letter in case a full fare basis code is passed
	letter := string(cleaned)

	if cabin, exists := cabinMap[letter]; exists {
		return cabin
	}

	return "unknown"
}
