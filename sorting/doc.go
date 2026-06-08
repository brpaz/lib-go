// Package sorting provides types for validated ORDER BY clauses in repository queries.
//
// [NewSort] validates the sort field against a caller-supplied allow-list, preventing
// SQL injection when field names originate from user input.
//
// # Usage
//
// Define the allowed fields once per repository, then parse user input:
//
//	var allowedSortFields = []string{"created_at", "name", "email"}
//
//	func (r *UserRepo) List(ctx context.Context, pager pagination.OffsetPager, s sorting.Sort) ([]*User, error) {
//	    s = sorting.Default(s, "created_at", sorting.DESC)
//	    var users []User
//	    err := r.db.WithContext(ctx).
//	        Order(s.SQL()).
//	        Offset(pager.Offset()).
//	        Limit(pager.Limit()).
//	        Find(&users).Error
//	    return users, err
//	}
//
// Parse sort parameters from an HTTP request and pass them down:
//
//	func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
//	    field := r.URL.Query().Get("sort_by")   // e.g. "name"
//	    dir   := r.URL.Query().Get("sort_dir")  // e.g. "asc"
//
//	    s, err := sorting.NewSort(field, dir, allowedSortFields)
//	    if err != nil {
//	        // field not in allow-list or direction invalid — return 400
//	    }
//
//	    users, err := h.repo.List(ctx, pager, s)
//	    // ...
//	}
//
// When no sort parameters are provided, pass a zero [Sort] and rely on [Default]:
//
//	s, _ := sorting.NewSort("", "ASC", allowedSortFields) // error — use zero value instead
//	s    := sorting.Sort{}                                 // zero value signals "no preference"
//	s     = sorting.Default(s, "created_at", sorting.DESC) // repo applies fallback
//
// # Multi-column sorting
//
// Build a [Sorts] from one [NewSort] call per ORDER BY column — each is
// validated against the same allow-list — then join them with [Sorts.SQL]:
//
//	var sorts sorting.Sorts
//	for _, p := range parsedSortParams { // e.g. [{"status","asc"}, {"created_at","desc"}]
//	    s, err := sorting.NewSort(p.Field, p.Dir, allowedSortFields)
//	    if err != nil {
//	        // invalid field or direction — return 400
//	    }
//	    sorts = append(sorts, s)
//	}
//
//	r.db.WithContext(ctx).Order(sorts.SQL()) // "status ASC, created_at DESC"
//
// How sort parameters are encoded on the wire (repeated query params, a
// comma-separated list, JSON:API-style "-field" prefixes, ...) is an API
// contract decision left to the caller — this package only validates and
// renders already-parsed (field, direction) pairs.
package sorting
