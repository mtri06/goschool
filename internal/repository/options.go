package repository

import (
	"fmt"
	"goschool/pkg/model"
	"strings"
)

func paginationToSQL(p model.Pagination) string {
	offset := (p.Page - 1) * p.PageSize
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.PageSize, offset)
}

func orderByToSQL(o model.OrderBy) string {
	if len(o) == 0 {
		return ""
	}

	var orderExps []string
	for _, order := range o {
		if order.Field == "" {
			continue
		}

		exp := quoteField(order.Field)
		if order.Desc {
			exp += " DESC"
		}

		orderExps = append(orderExps, exp)
	}

	sql := " ORDER BY " + strings.Join(orderExps, ", ")
	return sql
}

// quoteIdent safely quotes an SQL identifier (e.g. column or table name) to prevent SQL injection through identifiers.
func quoteIdent(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func quoteField(field string) string {
	parts := strings.SplitN(field, ".", 2)
	if len(parts) == 2 {
		return fmt.Sprintf("%s.%s", quoteIdent(parts[0]), quoteIdent(parts[1]))
	}
	return quoteIdent(parts[0])
}
