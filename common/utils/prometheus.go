package utils

import (
	"regexp"
	"strconv"
)

type PrometheusMatcher struct {
	Label    string
	Value    string
	Operator string
}

func (q *PrometheusMatcher) String() string {
	return q.Label + q.Operator + strconv.Quote(q.Value)
}

func NewPrometheusMatcher(l, op, v string) PrometheusMatcher {
	return PrometheusMatcher{
		Label:    l,
		Value:    v,
		Operator: op,
	}
}

func NewRegQuoteMetaMatcher(key string, value string) PrometheusMatcher {
	return PrometheusMatcher{
		Label:    key,
		Value:    regexp.QuoteMeta(value),
		Operator: `=~`,
	}
}

func NewRegQuoteMetaMatchers(key string, values []string) PrometheusMatcher {
	var v string
	for idx, value := range values {
		v += regexp.QuoteMeta(value)
		if idx < len(values)-1 {
			v += "|"
		}
	}

	return PrometheusMatcher{
		Label:    key,
		Value:    v,
		Operator: `=~`,
	}
}

type QueryString struct {
	MetricName string
	Matchers   []PrometheusMatcher
}

func (q *QueryString) String() string {
	res := q.MetricName
	if len(q.Matchers) != 0 {
		res += "{"
		for idx, matcher := range q.Matchers {
			res += matcher.String()
			if idx < len(q.Matchers)-1 {
				res += ", "
			}
		}
		res += "}"
	}

	return res
}
