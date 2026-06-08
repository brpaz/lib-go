// Package pagination provides types and helpers for offset-based and
// cursor-based pagination.
//
// # Offset pagination
//
// [OffsetPager] carries page/pageSize input for SQL OFFSET/LIMIT queries.
// [NewOffsetPager] clamps both values to safe ranges (page ≥ 1, 1 ≤ pageSize ≤
// [MaxPageSize] by default). Override the bounds per call site with
// [WithDefaultPageSize] / [WithMaxPageSize] — e.g. a higher cap for an
// internal API: pagination.NewOffsetPager(page, pageSize, pagination.WithMaxPageSize(500)).
// [Page] wraps the result set with total-count metadata:
//
//	pager := pagination.NewOffsetPager(page, pageSize) // from query params
//
//	var rows []User
//	var total int64
//	db.Model(&User{}).Count(&total).
//	    Offset(pager.Offset()).Limit(pager.Limit()).Find(&rows)
//
//	result := pagination.NewPage(rows, total, pager)
//	// result.TotalPages(), result.HasNext(), result.HasPrev()
//
// # Cursor pagination
//
// [CursorPager] carries an opaque cursor token and a limit for keyset-style
// queries. [NewCursorPager] clamps the limit the same way [NewOffsetPager]
// clamps page size — override with [WithDefaultLimit] / [WithMaxLimit].
// A cursor is one or more base64-encoded raw values — one per ORDER BY
// column, in column order — so multi-column sorts (e.g. created_at, id)
// stay stable across pages without skipping or repeating rows on ties.
// [CursorPage] wraps the result set with encoded next/prev cursor tokens:
//
//	pager := pagination.NewCursorPager(cursorParam, limitParam)
//
//	vals, err := pager.DecodedCursor() // []string{createdAt, id} for WHERE (created_at, id) > (?, ?)
//
//	// after fetching items, encode the boundary values in the same column order:
//	last, first := items[len(items)-1], items[0]
//	result := pagination.NewCursorPage(items,
//	    []string{last.CreatedAt.Format(time.RFC3339Nano), last.ID.String()},
//	    []string{first.CreatedAt.Format(time.RFC3339Nano), first.ID.String()},
//	)
//	// result.NextCursor, result.PrevCursor, result.HasNext, result.HasPrev
//
// A single-column sort works the same way with one-element slices —
// pagination.NewCursorPage(items, []string{last.ID.String()}, []string{first.ID.String()}).
package pagination
