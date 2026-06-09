package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/brpaz/lib-go/pagination"
	"github.com/brpaz/lib-go/sorting"
)

// ParseSort parses the ?sort query parameter into a [sorting.Sorts] slice.
//
// Format: comma-separated field names; prefix "-" for descending order.
//
//	?sort=-created_at,name  →  ORDER BY created_at DESC, full_name ASC
//
// allowed maps API field names to domain field names, e.g.:
//
//	map[string]string{"createdAt": "created_at", "name": "full_name"}
//
// Returns nil, nil when the parameter is absent.
// Returns an error for unknown fields or invalid input.
func ParseSort(r *http.Request, allowed map[string]string) (sorting.Sorts, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("sort"))
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	result := make(sorting.Sorts, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		dir := sorting.ASC
		field := part
		if strings.HasPrefix(part, "-") {
			dir = sorting.DESC
			field = part[1:]
		}

		if field == "" {
			return nil, fmt.Errorf("invalid sort field %q", part)
		}

		domainField, ok := allowed[field]
		if !ok {
			return nil, fmt.Errorf("invalid sort field %q", field)
		}

		result = append(result, sorting.Sort{Field: domainField, Dir: dir})
	}

	return result, nil
}

// ParseOffsetPager parses ?page and ?page_size query parameters into a
// [pagination.OffsetPager]. Invalid or absent values fall back to package
// defaults. Pass [pagination.WithDefaultPageSize] or [pagination.WithMaxPageSize]
// to override bounds.
//
//	?page=2&page_size=50  →  OffsetPager{Page:2, PageSize:50}
func ParseOffsetPager(r *http.Request, opts ...pagination.OffsetOption) (pagination.OffsetPager, error) {
	page, err := queryInt(r, "page")
	if err != nil {
		return pagination.OffsetPager{}, fmt.Errorf("invalid page: %w", err)
	}
	pageSize, err := queryInt(r, "page_size")
	if err != nil {
		return pagination.OffsetPager{}, fmt.Errorf("invalid page_size: %w", err)
	}
	return pagination.NewOffsetPager(page, pageSize, opts...), nil
}

// ParseCursorPager parses ?cursor and ?limit query parameters into a
// [pagination.CursorPager]. Invalid or absent values fall back to package
// defaults. Pass [pagination.WithDefaultLimit] or [pagination.WithMaxLimit]
// to override bounds.
//
//	?cursor=xxx&limit=20  →  CursorPager{Cursor:"xxx", Limit:20}
func ParseCursorPager(r *http.Request, opts ...pagination.CursorOption) (pagination.CursorPager, error) {
	cursor := r.URL.Query().Get("cursor")
	limit, err := queryInt(r, "limit")
	if err != nil {
		return pagination.CursorPager{}, fmt.Errorf("invalid limit: %w", err)
	}
	return pagination.NewCursorPager(cursor, limit, opts...), nil
}

// queryInt returns the integer value of a query parameter.
// Returns 0 (not an error) when the parameter is absent or empty.
// Returns an error when the value is present but not a valid integer.
func queryInt(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid integer", raw)
	}
	return v, nil
}
