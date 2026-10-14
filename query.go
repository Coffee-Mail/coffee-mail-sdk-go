package coffeemail

import (
	"net/url"
	"strconv"
	"time"
)

type queryBuilder url.Values

func newQuery() queryBuilder {
	return queryBuilder{}
}

func (q queryBuilder) str(name, value string) queryBuilder {
	if value != "" {
		url.Values(q).Set(name, value)
	}
	return q
}

func (q queryBuilder) num(name string, value int) queryBuilder {
	if value > 0 {
		url.Values(q).Set(name, strconv.Itoa(value))
	}
	return q
}

func (q queryBuilder) time(name string, value time.Time) queryBuilder {
	if !value.IsZero() {
		url.Values(q).Set(name, value.UTC().Format(time.RFC3339))
	}
	return q
}

func (q queryBuilder) values() url.Values {
	return url.Values(q)
}
