package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/pagination"
)

func TestNewOffsetPager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{name: "valid values", page: 2, pageSize: 10, wantPage: 2, wantPageSize: 10},
		{name: "page zero defaults to 1", page: 0, pageSize: 10, wantPage: 1, wantPageSize: 10},
		{
			name:         "negative page defaults to 1",
			page:         -5,
			pageSize:     10,
			wantPage:     1,
			wantPageSize: 10,
		},
		{name: "pageSize zero defaults to 20", page: 1, pageSize: 0, wantPage: 1, wantPageSize: 20},
		{
			name:         "negative pageSize defaults to 20",
			page:         1,
			pageSize:     -1,
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "pageSize above max clamped to 100",
			page:         1,
			pageSize:     200,
			wantPage:     1,
			wantPageSize: 100,
		},
		{
			name:         "pageSize exactly max allowed",
			page:         1,
			pageSize:     100,
			wantPage:     1,
			wantPageSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := pagination.NewOffsetPager(tt.page, tt.pageSize)

			assert.Equal(t, tt.wantPage, p.Page)
			assert.Equal(t, tt.wantPageSize, p.PageSize)
		})
	}
}

func TestNewOffsetPager_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithDefaultPageSize overrides fallback", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewOffsetPager(1, 0, pagination.WithDefaultPageSize(50))
		assert.Equal(t, 50, p.PageSize)
	})

	t.Run("WithMaxPageSize overrides cap", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewOffsetPager(1, 1000, pagination.WithMaxPageSize(500))
		assert.Equal(t, 500, p.PageSize)
	})

	t.Run("WithMaxPageSize below default still clamps the fallback", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewOffsetPager(1, 0, pagination.WithMaxPageSize(10))
		assert.Equal(t, 10, p.PageSize)
	})

	t.Run("options compose", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewOffsetPager(1, 0,
			pagination.WithDefaultPageSize(200),
			pagination.WithMaxPageSize(300),
		)
		assert.Equal(t, 200, p.PageSize)
	})
}

func TestOffsetPager_Offset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{name: "first page", page: 1, pageSize: 20, want: 0},
		{name: "second page", page: 2, pageSize: 20, want: 20},
		{name: "third page with small size", page: 3, pageSize: 5, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := pagination.NewOffsetPager(tt.page, tt.pageSize)
			assert.Equal(t, tt.want, p.Offset())
		})
	}
}

func TestOffsetPager_Limit(t *testing.T) {
	t.Parallel()

	p := pagination.NewOffsetPager(1, 15)
	assert.Equal(t, 15, p.Limit())
}

func TestPage_TotalPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		total    int64
		pageSize int
		want     int
	}{
		{name: "exact division", total: 40, pageSize: 20, want: 2},
		{name: "remainder rounds up", total: 41, pageSize: 20, want: 3},
		{name: "total smaller than page size", total: 5, pageSize: 20, want: 1},
		{name: "zero total", total: 0, pageSize: 20, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pager := pagination.NewOffsetPager(1, tt.pageSize)
			page := pagination.NewPage([]string{}, tt.total, pager)
			assert.Equal(t, tt.want, page.TotalPages())
		})
	}

	t.Run("zero page size returns 0", func(t *testing.T) {
		t.Parallel()

		page := pagination.Page[string]{Total: 10, PageSize: 0}
		assert.Equal(t, 0, page.TotalPages())
	})
}

func TestPage_HasNext(t *testing.T) {
	t.Parallel()

	t.Run("has next when not on last page", func(t *testing.T) {
		t.Parallel()

		pager := pagination.NewOffsetPager(1, 10)
		page := pagination.NewPage([]string{}, 25, pager)
		assert.True(t, page.HasNext())
	})

	t.Run("no next on last page", func(t *testing.T) {
		t.Parallel()

		pager := pagination.NewOffsetPager(3, 10)
		page := pagination.NewPage([]string{}, 25, pager)
		assert.False(t, page.HasNext())
	})

	t.Run("no next when total fits in one page", func(t *testing.T) {
		t.Parallel()

		pager := pagination.NewOffsetPager(1, 20)
		page := pagination.NewPage([]string{}, 5, pager)
		assert.False(t, page.HasNext())
	})
}

func TestPage_HasPrev(t *testing.T) {
	t.Parallel()

	t.Run("no prev on first page", func(t *testing.T) {
		t.Parallel()

		pager := pagination.NewOffsetPager(1, 10)
		page := pagination.NewPage([]string{}, 50, pager)
		assert.False(t, page.HasPrev())
	})

	t.Run("has prev on page 2", func(t *testing.T) {
		t.Parallel()

		pager := pagination.NewOffsetPager(2, 10)
		page := pagination.NewPage([]string{}, 50, pager)
		assert.True(t, page.HasPrev())
	})
}

func TestNewPage(t *testing.T) {
	t.Parallel()

	pager := pagination.NewOffsetPager(2, 10)
	items := []string{"a", "b", "c"}
	page := pagination.NewPage(items, 50, pager)

	assert.Equal(t, items, page.Items)
	assert.Equal(t, int64(50), page.Total)
	assert.Equal(t, 2, page.PageNumber)
	assert.Equal(t, 10, page.PageSize)
}
