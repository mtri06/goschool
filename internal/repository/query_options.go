package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Operator string

const (
	OpEquals          Operator = "="
	OpNotEquals       Operator = "!="
	OpGreaterThan     Operator = ">"
	OpLessThan        Operator = "<"
	OpGreaterEqual    Operator = ">="
	OpLessEqual       Operator = "<="
	OpLike            Operator = "LIKE"
	OpLikeInsensitive Operator = "ILIKE"
	OpIn              Operator = "IN"
)

var allowedOperators = map[Operator]struct{}{
	OpEquals:          {},
	OpNotEquals:       {},
	OpGreaterThan:     {},
	OpLessThan:        {},
	OpGreaterEqual:    {},
	OpLessEqual:       {},
	OpLike:            {},
	OpLikeInsensitive: {},
	OpIn:              {},
}

type Filter struct {
	Field string
	Op    Operator
	Value any

	alias string
}

func NewFilter(field string, op Operator, value any) *Filter {
	return &Filter{Field: field, Op: op, Value: value}
}

type Filters []*Filter

func (f Filters) toWhereClause() (where string, args []any, err error) {
	if len(f) == 0 {
		return "", nil, nil
	}

	args = make([]any, 0, len(f))
	conditions := make([]string, 0, len(f))
	for _, item := range f {
		if item == nil {
			continue
		}
		if item.Field == "" {
			return "", nil, fmt.Errorf("filter item field cannot be empty")
		}
		if _, ok := allowedOperators[item.Op]; !ok {
			return "", nil, fmt.Errorf("invalid operator: %s", item.Op)
		}

		field := item.Field
		if item.alias != "" {
			field = fmt.Sprintf("%s.%s", item.alias, field)
		}

		if item.Op == OpIn {
			conditions = append(conditions, fmt.Sprintf("%s IN (?)", field))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s %s ?", field, item.Op))
		}

		args = append(args, item.Value)
	}
	where = "WHERE " + strings.Join(conditions, " AND ")
	where, args, err = sqlx.In(where, args...)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build filter query: %w", err)
	}

	where = sqlx.Rebind(sqlx.DOLLAR, where)

	return where, args, nil
}

func (f Filters) setAlias(alias string) {
	for _, item := range f {
		if item != nil {
			item.alias = alias
		}
	}
}

type Pagination struct {
	Page     int
	PageSize int
}

func (p *Pagination) toLimitOffset() string {
	if p == nil {
		return ""
	}
	offset := (p.Page - 1) * p.PageSize
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.PageSize, offset)
}
