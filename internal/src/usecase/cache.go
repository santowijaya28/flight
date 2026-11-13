package usecase

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/flight/internal/src/model"
)

func buildCacheKey(in model.SearchInput) (string, error) {
	normalized := in

	normalized.Origin = strings.ToUpper(strings.TrimSpace(in.Origin))
	normalized.Destination = strings.ToUpper(strings.TrimSpace(in.Destination))

	normalized.CabinClass = strings.ToLower(strings.TrimSpace(in.CabinClass))
	normalized.Sort = strings.ToLower(strings.TrimSpace(in.Sort))

	normalized.DepartureFrom = strings.TrimSpace(in.DepartureFrom)
	normalized.DepartureTo = strings.TrimSpace(in.DepartureTo)

	normalized.ArrivalFrom = strings.TrimSpace(in.ArrivalFrom)
	normalized.ArrivalTo = strings.TrimSpace(in.ArrivalTo)

	normalized.Airlines = make([]string, 0, len(in.Airlines))
	for _, airline := range in.Airlines {
		airline = strings.ToLower(strings.TrimSpace(airline))
		if airline != "" {
			normalized.Airlines = append(normalized.Airlines, airline)
		}
	}
	sort.Strings(normalized.Airlines)

	data, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	return "search:v1:" + string(data), nil
}
