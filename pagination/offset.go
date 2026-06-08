package pagination

import "math"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// OffsetPager holds offset-based pagination input for repository queries.
type OffsetPager struct {
	Page     int
	PageSize int
}

// offsetLimits holds the page-size bounds NewOffsetPager clamps against.
type offsetLimits struct {
	defaultPageSize int
	maxPageSize     int
}

// OffsetOption configures the bounds NewOffsetPager clamps page size against.
type OffsetOption func(*offsetLimits)

// WithDefaultPageSize overrides the page size used when the caller passes a
// value below 1 (e.g. an absent query parameter). Defaults to [DefaultPageSize].
func WithDefaultPageSize(n int) OffsetOption {
	return func(l *offsetLimits) { l.defaultPageSize = n }
}

// WithMaxPageSize overrides the upper bound page size is clamped to.
// Defaults to [MaxPageSize].
func WithMaxPageSize(n int) OffsetOption {
	return func(l *offsetLimits) { l.maxPageSize = n }
}

// NewOffsetPager creates an OffsetPager, clamping values to valid ranges.
// By default page defaults to [DefaultPage], pageSize to [DefaultPageSize],
// and is capped at [MaxPageSize]. Override either bound with [WithDefaultPageSize]
// or [WithMaxPageSize] — e.g. to allow larger pages on an internal API:
//
//	pager := pagination.NewOffsetPager(page, pageSize, pagination.WithMaxPageSize(500))
func NewOffsetPager(page, pageSize int, opts ...OffsetOption) OffsetPager {
	limits := offsetLimits{
		defaultPageSize: DefaultPageSize,
		maxPageSize:     MaxPageSize,
	}
	for _, opt := range opts {
		opt(&limits)
	}

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = limits.defaultPageSize
	}
	if pageSize > limits.maxPageSize {
		pageSize = limits.maxPageSize
	}
	return OffsetPager{Page: page, PageSize: pageSize}
}

// Offset returns the DB offset for this page.
func (p OffsetPager) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the DB limit (same as PageSize).
func (p OffsetPager) Limit() int {
	return p.PageSize
}

// Page is a paginated result set for offset-based pagination.
type Page[T any] struct {
	Items      []T
	Total      int64
	PageNumber int
	PageSize   int
}

// NewPage creates a Page from query results and the originating OffsetPager.
func NewPage[T any](items []T, total int64, pager OffsetPager) Page[T] {
	return Page[T]{
		Items:      items,
		Total:      total,
		PageNumber: pager.Page,
		PageSize:   pager.PageSize,
	}
}

// TotalPages returns the number of pages for the total count.
func (p Page[T]) TotalPages() int {
	if p.PageSize == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}

// HasNext reports whether a next page exists.
func (p Page[T]) HasNext() bool {
	return p.PageNumber < p.TotalPages()
}

// HasPrev reports whether a previous page exists.
func (p Page[T]) HasPrev() bool {
	return p.PageNumber > 1
}
