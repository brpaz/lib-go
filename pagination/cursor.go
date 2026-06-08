package pagination

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const DefaultCursorLimit = 20

// cursorSeparator joins multi-value cursors before encoding. It is the ASCII
// Unit Separator — a control character that won't appear in IDs, timestamps,
// or other typical cursor values.
const cursorSeparator = "\x1f"

// CursorPager holds cursor-based pagination input for repository queries.
type CursorPager struct {
	// Cursor is an opaque token pointing to the last seen item. Empty means start from beginning.
	Cursor string
	Limit  int
}

// cursorLimits holds the limit bounds NewCursorPager clamps against.
type cursorLimits struct {
	defaultLimit int
	maxLimit     int
}

// CursorOption configures the bounds NewCursorPager clamps its limit against.
type CursorOption func(*cursorLimits)

// WithDefaultLimit overrides the limit used when the caller passes a value
// below 1 (e.g. an absent query parameter). Defaults to [DefaultCursorLimit].
func WithDefaultLimit(n int) CursorOption {
	return func(l *cursorLimits) { l.defaultLimit = n }
}

// WithMaxLimit overrides the upper bound the limit is clamped to.
// Defaults to [MaxPageSize].
func WithMaxLimit(n int) CursorOption {
	return func(l *cursorLimits) { l.maxLimit = n }
}

// NewCursorPager creates a CursorPager with a valid limit. By default the
// limit defaults to [DefaultCursorLimit] and is capped at [MaxPageSize].
// Override either bound with [WithDefaultLimit] or [WithMaxLimit] — e.g. to
// allow larger pages on an internal API:
//
//	pager := pagination.NewCursorPager(cursor, limit, pagination.WithMaxLimit(500))
func NewCursorPager(cursor string, limit int, opts ...CursorOption) CursorPager {
	limits := cursorLimits{
		defaultLimit: DefaultCursorLimit,
		maxLimit:     MaxPageSize,
	}

	for _, opt := range opts {
		opt(&limits)
	}

	if limit < 1 {
		limit = limits.defaultLimit
	}
	if limit > limits.maxLimit {
		limit = limits.maxLimit
	}
	return CursorPager{Cursor: cursor, Limit: limit}
}

// DecodedCursor decodes the opaque cursor to its raw values (e.g. the
// last-seen sort-key values used to keep a multi-column ORDER BY stable, such
// as []string{createdAt, id}). Returns nil if cursor is empty (first page).
func (p CursorPager) DecodedCursor() ([]string, error) {
	if p.Cursor == "" {
		return nil, nil
	}
	return DecodeCursor(p.Cursor)
}

// CursorPage is a cursor-paginated result set.
type CursorPage[T any] struct {
	Items      []T
	NextCursor string
	PrevCursor string
	HasNext    bool
	HasPrev    bool
}

// NewCursorPage builds a CursorPage.
// nextVals and prevVals are the raw cursor values of the last/first item —
// one per ORDER BY column, in the same order (e.g. []string{createdAt, id}).
// Pass nil (or empty) for nextVals/prevVals if there is no next/prev page.
func NewCursorPage[T any](items []T, nextVals, prevVals []string) CursorPage[T] {
	p := CursorPage[T]{Items: items}

	if len(nextVals) > 0 {
		p.NextCursor = EncodeCursor(nextVals...)
		p.HasNext = true
	}
	if len(prevVals) > 0 {
		p.PrevCursor = EncodeCursor(prevVals...)
		p.HasPrev = true
	}

	return p
}

// EncodeCursor encodes one or more raw values into a single opaque cursor
// token. Pass one value per ORDER BY column, in column order, to keep
// multi-column sorts stable across pages (e.g. EncodeCursor(createdAt, id)).
func EncodeCursor(values ...string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(values, cursorSeparator)))
}

// DecodeCursor decodes an opaque cursor token back to its raw values, in the
// same order they were passed to EncodeCursor.
func DecodeCursor(cursor string) ([]string, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	return strings.Split(string(b), cursorSeparator), nil
}
