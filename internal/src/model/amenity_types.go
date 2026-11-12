package model

import (
	"sort"
	"strings"
)

// NormalizeAmenities trims, lowercases, deduplicates, and sorts amenities.
// It always returns a non-nil slice so JSON encodes empty amenities as [].
func NormalizeAmenities(values []string) []string {
	amenities := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		amenity := strings.ToLower(strings.TrimSpace(value))
		amenity = strings.ReplaceAll(amenity, "_", " ")
		if amenity == "" {
			continue
		}

		if _, exists := seen[amenity]; exists {
			continue
		}

		seen[amenity] = struct{}{}
		amenities = append(amenities, amenity)
	}

	sort.Strings(amenities)
	return amenities
}
