package sorting

import (
	"fmt"
	"slices"
	"strings"
)

// Direction is a sort direction: ASC or DESC.
type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

// ParseDirection parses a case-insensitive string into a Direction.
func ParseDirection(s string) (Direction, error) {
	switch strings.ToUpper(s) {
	case "ASC":
		return ASC, nil
	case "DESC":
		return DESC, nil
	default:
		return "", fmt.Errorf("invalid sort direction %q: must be ASC or DESC", s)
	}
}

// Sort holds a validated field and direction for ORDER BY clauses.
type Sort struct {
	Field string
	Dir   Direction
}

// NewSort creates a Sort after validating field against allowed and parsing dir.
// Returns an error if field is not in the allow-list or dir is not ASC/DESC.
// The allow-list prevents SQL injection when field names come from user input.
func NewSort(field, dir string, allowed []string) (Sort, error) {
	if !slices.Contains(allowed, field) {
		return Sort{}, fmt.Errorf("invalid sort field %q", field)
	}

	d, err := ParseDirection(dir)
	if err != nil {
		return Sort{}, err
	}

	return Sort{Field: field, Dir: d}, nil
}

// Default returns s if its Field is non-empty, otherwise a Sort with defaultField and defaultDir.
func Default(s Sort, defaultField string, defaultDir Direction) Sort {
	if s.Field != "" {
		return s
	}
	return Sort{Field: defaultField, Dir: defaultDir}
}

// SQL returns the ORDER BY fragment (e.g. "created_at DESC") safe for direct use in queries.
// Both Field and Dir are validated at construction, so this output is injection-safe.
func (s Sort) SQL() string {
	return s.Field + " " + string(s.Dir)
}

// Sorts is an ordered list of validated Sort clauses for multi-column ORDER BY.
type Sorts []Sort

// SQL returns the joined ORDER BY fragment (e.g. "status ASC, created_at DESC")
// safe for direct use in queries. Each element is validated at construction via
// [NewSort], so this output is injection-safe.
func (s Sorts) SQL() string {
	clauses := make([]string, len(s))
	for i, sort := range s {
		clauses[i] = sort.SQL()
	}
	return strings.Join(clauses, ", ")
}
