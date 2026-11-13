package model

import (
	"math"
	"strconv"
	"strings"
)

func FormatIDR(amount float64) string {
	return "Rp " + formatThousands(int64(math.Round(amount)))
}

func formatThousands(value int64) string {
	if value == 0 {
		return "0"
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	digits := strconv.FormatInt(value, 10)
	var builder strings.Builder
	builder.Grow(len(digits) + len(digits)/3)

	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			builder.WriteByte('.')
		}
		builder.WriteRune(digit)
	}

	return sign + builder.String()
}
