package handler

import (
	"goschool/pkg/model"
	"strings"
)

func parseOrderBy(order []string) model.OrderBy {
	var orderBy model.OrderBy
	for _, ob := range order {
		ob, desc := strings.CutPrefix(ob, "-")
		orderBy = append(orderBy, model.Order{Field: ob, Desc: desc})
	}
	return orderBy
}
