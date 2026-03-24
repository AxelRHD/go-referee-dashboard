package validation

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func requireField(form url.Values, field, label string, errors map[string]string) string {
	val := strings.TrimSpace(form.Get(field))
	if val == "" {
		errors[field] = fmt.Sprintf("%s ist ein Pflichtfeld.", label)
	}
	return val
}

func parseIntField(form url.Values, field, label string, errors map[string]string, defaultVal int64) int64 {
	raw := strings.TrimSpace(form.Get(field))
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		errors[field] = fmt.Sprintf("%s muss eine ganze Zahl sein.", label)
		return defaultVal
	}
	if v < 0 {
		errors[field] = fmt.Sprintf("%s darf nicht negativ sein.", label)
		return defaultVal
	}
	return v
}

func parseFloatField(form url.Values, field, label string, errors map[string]string, defaultVal float64) float64 {
	raw := strings.TrimSpace(form.Get(field))
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		errors[field] = fmt.Sprintf("%s muss eine Zahl sein.", label)
		return defaultVal
	}
	if v < 0 {
		errors[field] = fmt.Sprintf("%s darf nicht negativ sein.", label)
		return defaultVal
	}
	return v
}

func optionalField(form url.Values, field string) string {
	return strings.TrimSpace(form.Get(field))
}
