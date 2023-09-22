package utils

import (
	"regexp"
)

func RegQuoteOr(values []string) string {
	var v string
	for idx, value := range values {
		v += regexp.QuoteMeta(value)
		if idx < len(values)-1 {
			v += "|"
		}
	}

	return v
}
