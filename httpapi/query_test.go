package httpapi_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/httpapi"
	"github.com/brpaz/lib-go/pagination"
	"github.com/brpaz/lib-go/sorting"
)

var testSortAllowed = map[string]string{
	"name":       "full_name",
	"createdAt":  "created_at",
	"totalPrice": "total_price",
}

func TestParseSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawQuery string
		want     sorting.Sorts
		wantErr  string
	}{
		{
			name:     "absent param returns nil",
			rawQuery: "",
			want:     nil,
		},
		{
			name:     "single asc",
			rawQuery: "sort=name",
			want:     sorting.Sorts{{Field: "full_name", Dir: sorting.ASC}},
		},
		{
			name:     "single desc",
			rawQuery: "sort=-createdAt",
			want:     sorting.Sorts{{Field: "created_at", Dir: sorting.DESC}},
		},
		{
			name:     "multi field",
			rawQuery: "sort=-createdAt,name",
			want: sorting.Sorts{
				{Field: "created_at", Dir: sorting.DESC},
				{Field: "full_name", Dir: sorting.ASC},
			},
		},
		{
			name:     "api name maps to domain name",
			rawQuery: "sort=totalPrice",
			want:     sorting.Sorts{{Field: "total_price", Dir: sorting.ASC}},
		},
		{
			name:     "unknown field returns error",
			rawQuery: "sort=unknown",
			wantErr:  `invalid sort field "unknown"`,
		},
		{
			name:     "bare dash returns error",
			rawQuery: "sort=-",
			wantErr:  `invalid sort field "-"`,
		},
		{
			name:     "empty segment is skipped",
			rawQuery: "sort=,name",
			want:     sorting.Sorts{{Field: "full_name", Dir: sorting.ASC}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest("GET", "/", nil)
			r.URL.RawQuery = tc.rawQuery

			got, err := httpapi.ParseSort(r, testSortAllowed)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseOffsetPager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawQuery string
		opts     []pagination.OffsetOption
		want     pagination.OffsetPager
		wantErr  string
	}{
		{
			name:     "absent params use defaults",
			rawQuery: "",
			want:     pagination.OffsetPager{Page: 1, PageSize: pagination.DefaultPageSize},
		},
		{
			name:     "explicit page and page_size",
			rawQuery: "page=3&page_size=10",
			want:     pagination.OffsetPager{Page: 3, PageSize: 10},
		},
		{
			name:     "page_size clamped to max",
			rawQuery: "page=1&page_size=999",
			want:     pagination.OffsetPager{Page: 1, PageSize: pagination.MaxPageSize},
		},
		{
			name:     "custom max via option",
			rawQuery: "page=1&page_size=200",
			opts:     []pagination.OffsetOption{pagination.WithMaxPageSize(500)},
			want:     pagination.OffsetPager{Page: 1, PageSize: 200},
		},
		{
			name:     "invalid page returns error",
			rawQuery: "page=abc",
			wantErr:  "invalid page",
		},
		{
			name:     "invalid page_size returns error",
			rawQuery: "page_size=abc",
			wantErr:  "invalid page_size",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest("GET", "/", nil)
			r.URL.RawQuery = tc.rawQuery

			got, err := httpapi.ParseOffsetPager(r, tc.opts...)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseCursorPager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawQuery string
		opts     []pagination.CursorOption
		want     pagination.CursorPager
		wantErr  string
	}{
		{
			name:     "absent params use defaults",
			rawQuery: "",
			want:     pagination.CursorPager{Cursor: "", Limit: pagination.DefaultCursorLimit},
		},
		{
			name:     "explicit cursor and limit",
			rawQuery: "cursor=abc123&limit=10",
			want:     pagination.CursorPager{Cursor: "abc123", Limit: 10},
		},
		{
			name:     "limit clamped to max",
			rawQuery: "cursor=x&limit=999",
			want:     pagination.CursorPager{Cursor: "x", Limit: pagination.MaxPageSize},
		},
		{
			name:     "custom max via option",
			rawQuery: "cursor=x&limit=200",
			opts:     []pagination.CursorOption{pagination.WithMaxLimit(500)},
			want:     pagination.CursorPager{Cursor: "x", Limit: 200},
		},
		{
			name:     "invalid limit returns error",
			rawQuery: "limit=abc",
			wantErr:  "invalid limit",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest("GET", "/", nil)
			r.URL.RawQuery = tc.rawQuery

			got, err := httpapi.ParseCursorPager(r, tc.opts...)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
