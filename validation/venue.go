package validation

import (
	"net/url"
	"strconv"
	"strings"
)

type VenueData struct {
	City    string
	Stadium string
	Lat     float64
	Lon     float64
}

func ValidateVenue(form url.Values) (VenueData, map[string]string) {
	errors := make(map[string]string)

	city := requireField(form, "city", "Stadt", errors)
	stadium := optionalField(form, "stadium")

	var lat, lon float64
	if raw := strings.TrimSpace(form.Get("lat")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			lat = v
		}
	}
	if raw := strings.TrimSpace(form.Get("lon")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			lon = v
		}
	}

	data := VenueData{
		City:    city,
		Stadium: stadium,
		Lat:     lat,
		Lon:     lon,
	}

	return data, errors
}
