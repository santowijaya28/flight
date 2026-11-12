package model

import "strings"

func OptionalAircraft(model string) *Aircraft {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil
	}

	return &Aircraft{Model: model}
}
