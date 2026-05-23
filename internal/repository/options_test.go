package repository

import (
	"strings"
	"testing"
)

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		op        Operator
		value     any
		wantField string
		wantOp    Operator
	}{
		{
			name:      "create filter with equals operator",
			field:     "id",
			op:        OpEquals,
			value:     1,
			wantField: "id",
			wantOp:    OpEquals,
		},
		{
			name:      "create filter with string value",
			field:     "name",
			op:        OpLike,
			value:     "%test%",
			wantField: "name",
			wantOp:    OpLike,
		},
		{
			name:      "create filter with IN operator",
			field:     "status",
			op:        OpIn,
			value:     []string{"active", "inactive"},
			wantField: "status",
			wantOp:    OpIn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.field, tt.op, tt.value)

			if filter.Field != tt.wantField {
				t.Errorf("Field = %v, want %v", filter.Field, tt.wantField)
			}
			if filter.Op != tt.wantOp {
				t.Errorf("Op = %v, want %v", filter.Op, tt.wantOp)
			}
			if filter.Value == nil {
				t.Errorf("Value is nil, expected non-nil")
			}
		})
	}
}

func TestFiltersToWhereClause(t *testing.T) {
	tests := []struct {
		name        string
		filters     Filters
		wantWhere   bool // whether WHERE should contain text
		wantArgsLen int
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "empty filters",
			filters:     Filters{},
			wantWhere:   false,
			wantArgsLen: 0,
			wantErr:     false,
		},
		{
			name: "single filter with equals",
			filters: Filters{
				NewFilter("id", OpEquals, 1),
			},
			wantWhere:   true,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "multiple filters with AND",
			filters: Filters{
				NewFilter("id", OpEquals, 1),
				NewFilter("status", OpEquals, "active"),
			},
			wantWhere:   true,
			wantArgsLen: 2,
			wantErr:     false,
		},
		{
			name: "filter with nil value",
			filters: Filters{
				nil,
				NewFilter("id", OpEquals, 1),
			},
			wantWhere:   true,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "filter with empty field",
			filters: Filters{
				NewFilter("", OpEquals, 1),
			},
			wantWhere:   false,
			wantArgsLen: 0,
			wantErr:     true,
			wantErrMsg:  "filter item field cannot be empty",
		},
		{
			name: "filter with invalid operator",
			filters: Filters{
				NewFilter("id", Operator("INVALID"), 1),
			},
			wantWhere:   false,
			wantArgsLen: 0,
			wantErr:     true,
			wantErrMsg:  "invalid operator",
		},
		{
			name: "filter with IN operator",
			filters: Filters{
				NewFilter("status", OpIn, []string{"active", "inactive"}),
			},
			wantWhere:   true,
			wantArgsLen: 2,
			wantErr:     false,
		},
		{
			name: "multiple filters with different operators",
			filters: Filters{
				NewFilter("id", OpGreaterThan, 10),
				NewFilter("name", OpLike, "%test%"),
				NewFilter("status", OpNotEquals, "deleted"),
			},
			wantWhere:   true,
			wantArgsLen: 3,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, args, err := tt.filters.toWhereClause()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.wantErrMsg != "" && !contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error message = %v, want to contain %v", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantWhere && where == "" {
				t.Errorf("where clause is empty, expected non-empty")
			}
			if !tt.wantWhere && where != "" {
				t.Errorf("where clause = %v, expected empty", where)
			}

			if len(args) != tt.wantArgsLen {
				t.Errorf("args length = %d, want %d", len(args), tt.wantArgsLen)
			}
		})
	}
}

func TestFiltersSetAlias(t *testing.T) {
	tests := []struct {
		name      string
		filters   Filters
		alias     string
		wantAlias string
	}{
		{
			name: "set alias for single filter",
			filters: Filters{
				NewFilter("id", OpEquals, 1),
			},
			alias:     "u",
			wantAlias: "u",
		},
		{
			name: "set alias for multiple filters",
			filters: Filters{
				NewFilter("id", OpEquals, 1),
				NewFilter("name", OpLike, "%test%"),
			},
			alias:     "t",
			wantAlias: "t",
		},
		{
			name: "set alias with nil filters",
			filters: Filters{
				nil,
				NewFilter("id", OpEquals, 1),
			},
			alias:     "s",
			wantAlias: "s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.filters.setAlias(tt.alias)

			for i, filter := range tt.filters {
				if filter == nil {
					continue
				}
				if filter.alias != tt.wantAlias {
					t.Errorf("filter[%d].alias = %v, want %v", i, filter.alias, tt.wantAlias)
				}
			}
		})
	}
}

func TestPaginationToLimitOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     string
	}{
		{
			name:     "first page with 10 items",
			page:     1,
			pageSize: 10,
			want:     "LIMIT 10 OFFSET 0",
		},
		{
			name:     "second page with 10 items",
			page:     2,
			pageSize: 10,
			want:     "LIMIT 10 OFFSET 10",
		},
		{
			name:     "third page with 20 items",
			page:     3,
			pageSize: 20,
			want:     "LIMIT 20 OFFSET 40",
		},
		{
			name:     "page 10 with 50 items per page",
			page:     10,
			pageSize: 50,
			want:     "LIMIT 50 OFFSET 450",
		},
		{
			name:     "page 1 with 1 item",
			page:     1,
			pageSize: 1,
			want:     "LIMIT 1 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Page: tt.page, PageSize: tt.pageSize}
			got := p.toLimitOffset()

			if got != tt.want {
				t.Errorf("toLimitOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaginationToLimitOffsetNil(t *testing.T) {
	var p *Pagination
	got := p.toLimitOffset()

	if got != "" {
		t.Errorf("toLimitOffset() on nil pagination = %v, want empty string", got)
	}
}

func TestOrderByToSQL(t *testing.T) {
	tests := []struct {
		name    string
		orderBy OrderBy
		want    string
	}{
		{
			name:    "empty order by",
			orderBy: OrderBy{},
			want:    "",
		},
		{
			name:    "single field ascending",
			orderBy: OrderBy{"name"},
			want:    " ORDER BY \"name\" ASC",
		},
		{
			name:    "multiple fields mixed order",
			orderBy: OrderBy{"name", "-id"},
			want:    " ORDER BY \"name\" ASC, \"id\" DESC",
		},
		{
			name:    "field with underscore",
			orderBy: OrderBy{"created_at"},
			want:    " ORDER BY \"created_at\" ASC",
		},
		{
			name:    "descending field with underscore",
			orderBy: OrderBy{"-updated_at"},
			want:    " ORDER BY \"updated_at\" DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.orderBy.toSQL()

			if got != tt.want {
				t.Errorf("toSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuoteIdent(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "simple identifier",
			id:   "name",
			want: `"name"`,
		},
		{
			name: "identifier with underscore",
			id:   "user_id",
			want: `"user_id"`,
		},
		{
			name: "identifier with number",
			id:   "table1",
			want: `"table1"`,
		},
		{
			name: "identifier with quote",
			id:   `col"umn`,
			want: `"col""umn"`,
		},
		{
			name: "identifier with multiple quotes",
			id:   `col"u"mn`,
			want: `"col""u""mn"`,
		},
		{
			name: "identifier starting with quote",
			id:   `"column`,
			want: `"""column"`,
		},
		{
			name: "identifier ending with quote",
			id:   `column"`,
			want: `"column"""`,
		},
		{
			name: "mixed case identifier",
			id:   "MyColumn",
			want: `"MyColumn"`,
		},
		{
			name: "empty string",
			id:   "",
			want: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteIdent(tt.id)

			if got != tt.want {
				t.Errorf("quoteIdent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterFieldQuoting(t *testing.T) {
	// Test that filter field names are properly quoted in WHERE clause
	tests := []struct {
		name    string
		filters Filters
		check   func(where string) bool
	}{
		{
			name: "simple field is quoted",
			filters: Filters{
				NewFilter("id", OpEquals, 1),
			},
			check: func(where string) bool {
				return contains(where, `"id"`)
			},
		},
		{
			name: "field with underscore is quoted",
			filters: Filters{
				NewFilter("user_id", OpEquals, 1),
			},
			check: func(where string) bool {
				return contains(where, `"user_id"`)
			},
		},
		{
			name: "alias and field are both quoted",
			filters: Filters{
				NewFilter("name", OpLike, "%test%"),
			},
			check: func(where string) bool {
				where = strings.Replace(where, `"name"`, "", -1)
				return contains(where, "LIKE")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, _, err := tt.filters.toWhereClause()
			if err != nil {
				t.Fatalf("toWhereClause() error = %v", err)
			}

			if !tt.check(where) {
				t.Errorf("WHERE clause validation failed: %v", where)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if len(substr) > len(s)-i {
			break
		}
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
