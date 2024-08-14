package query

import (
	"fmt"
	"strings"
)

type V []interface{}

type Conds struct {
	Key    string
	Values V
	Skip   bool
}

//	conds, vals := query.Build([]query.Conds{
//		{
//			Key:    "sender_id = ? OR receiver_id = ?",
//			Values: query.V{accountId, accountId},
//			Skip:   false,
//		},
//	})
//
// Build([]Conds) will return a query string and a slice of values that can be used in the QueryContext method.
func Build(conds []Conds) (string, []interface{}) {
	if len(conds) == 0 {
		return "", nil
	}
	var query string
	var values []interface{}
	var limitIdx int
	for idx, cond := range conds {
		if cond.Skip {
			continue
		}
		if strings.Contains(cond.Key, "LIMIT") {
			limitIdx = idx
			continue
		}
		if len(cond.Values) > 0 && cond.Values[0] != "" {
			query += fmt.Sprintf("%s AND ", cond.Key)
			values = append(values, cond.Values...)
		}
	}

	if len(query) == 0 {
		return "", nil
	}
	query = query[:len(query)-5]
	if limitIdx > 0 {
		query += fmt.Sprintf(" %s", conds[limitIdx].Key)
		values = append(values, conds[limitIdx].Values...)
	}

	return ReplacePlaceholder(query), values
}

func ReplacePlaceholder(q string) string {
	parts := strings.Split(q, "?")
	for i := 0; i < len(parts)-1; i++ {
		parts[i] += fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, "")
}
