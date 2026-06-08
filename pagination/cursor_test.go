package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/pagination"
)

func TestNewCursorPager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cursor    string
		limit     int
		wantLimit int
	}{
		{name: "valid values", cursor: "abc", limit: 10, wantLimit: 10},
		{name: "empty cursor first page", cursor: "", limit: 10, wantLimit: 10},
		{name: "zero limit defaults to 20", cursor: "", limit: 0, wantLimit: 20},
		{name: "negative limit defaults to 20", cursor: "", limit: -1, wantLimit: 20},
		{name: "limit above max clamped to 100", cursor: "", limit: 200, wantLimit: 100},
		{name: "limit exactly max allowed", cursor: "", limit: 100, wantLimit: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := pagination.NewCursorPager(tt.cursor, tt.limit)

			assert.Equal(t, tt.cursor, p.Cursor)
			assert.Equal(t, tt.wantLimit, p.Limit)
		})
	}
}

func TestNewCursorPager_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithDefaultLimit overrides fallback", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("", 0, pagination.WithDefaultLimit(50))
		assert.Equal(t, 50, p.Limit)
	})

	t.Run("WithMaxLimit overrides cap", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("", 1000, pagination.WithMaxLimit(500))
		assert.Equal(t, 500, p.Limit)
	})

	t.Run("WithMaxLimit below default still clamps the fallback", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("", 0, pagination.WithMaxLimit(10))
		assert.Equal(t, 10, p.Limit)
	})

	t.Run("options compose", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("", 0,
			pagination.WithDefaultLimit(200),
			pagination.WithMaxLimit(300),
		)
		assert.Equal(t, 200, p.Limit)
	})
}

func TestCursorPager_DecodedCursor(t *testing.T) {
	t.Parallel()

	t.Run("empty cursor returns nil", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("", 10)
		got, err := p.DecodedCursor()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("valid single-value cursor decodes correctly", func(t *testing.T) {
		t.Parallel()

		encoded := pagination.EncodeCursor("some-uuid-value")
		p := pagination.NewCursorPager(encoded, 10)

		got, err := p.DecodedCursor()
		require.NoError(t, err)
		assert.Equal(t, []string{"some-uuid-value"}, got)
	})

	t.Run("valid composite cursor decodes correctly", func(t *testing.T) {
		t.Parallel()

		encoded := pagination.EncodeCursor("2024-01-01T00:00:00Z", "some-uuid-value")
		p := pagination.NewCursorPager(encoded, 10)

		got, err := p.DecodedCursor()
		require.NoError(t, err)
		assert.Equal(t, []string{"2024-01-01T00:00:00Z", "some-uuid-value"}, got)
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		t.Parallel()

		p := pagination.NewCursorPager("not-valid-base64!!!", 10)
		_, err := p.DecodedCursor()
		require.Error(t, err)
	})
}

func TestEncodeCursor_DecodeCursor(t *testing.T) {
	t.Parallel()

	t.Run("round-trip single value", func(t *testing.T) {
		t.Parallel()

		values := []string{
			"simple",
			"550e8400-e29b-41d4-a716-446655440000",
			"2024-01-01T00:00:00Z:some-id",
		}

		for _, v := range values {
			encoded := pagination.EncodeCursor(v)
			decoded, err := pagination.DecodeCursor(encoded)
			require.NoError(t, err)
			assert.Equal(t, []string{v}, decoded)
		}
	})

	t.Run("round-trip composite value", func(t *testing.T) {
		t.Parallel()

		values := []string{"2024-01-01T00:00:00Z", "550e8400-e29b-41d4-a716-446655440000"}

		encoded := pagination.EncodeCursor(values...)
		decoded, err := pagination.DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, values, decoded)
	})

	t.Run("decode invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := pagination.DecodeCursor("!!!not-base64!!!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor")
	})
}

func TestNewCursorPage(t *testing.T) {
	t.Parallel()

	t.Run("with next and prev", func(t *testing.T) {
		t.Parallel()

		items := []string{"a", "b", "c"}
		page := pagination.NewCursorPage(items, []string{"last-id"}, []string{"first-id"})

		assert.Equal(t, items, page.Items)
		assert.True(t, page.HasNext)
		assert.True(t, page.HasPrev)
		assert.NotEmpty(t, page.NextCursor)
		assert.NotEmpty(t, page.PrevCursor)
	})

	t.Run("first page has no prev", func(t *testing.T) {
		t.Parallel()

		page := pagination.NewCursorPage([]string{"a"}, []string{"last-id"}, nil)

		assert.True(t, page.HasNext)
		assert.False(t, page.HasPrev)
		assert.NotEmpty(t, page.NextCursor)
		assert.Empty(t, page.PrevCursor)
	})

	t.Run("last page has no next", func(t *testing.T) {
		t.Parallel()

		page := pagination.NewCursorPage([]string{"a"}, nil, []string{"first-id"})

		assert.False(t, page.HasNext)
		assert.True(t, page.HasPrev)
		assert.Empty(t, page.NextCursor)
		assert.NotEmpty(t, page.PrevCursor)
	})

	t.Run("single page has no next or prev", func(t *testing.T) {
		t.Parallel()

		page := pagination.NewCursorPage([]string{"a"}, nil, nil)

		assert.False(t, page.HasNext)
		assert.False(t, page.HasPrev)
		assert.Empty(t, page.NextCursor)
		assert.Empty(t, page.PrevCursor)
	})

	t.Run("next cursor is encoded value of nextVals", func(t *testing.T) {
		t.Parallel()

		page := pagination.NewCursorPage([]string{}, []string{"my-id"}, nil)
		decoded, err := pagination.DecodeCursor(page.NextCursor)
		require.NoError(t, err)
		assert.Equal(t, []string{"my-id"}, decoded)
	})

	t.Run("composite next cursor preserves column order", func(t *testing.T) {
		t.Parallel()

		page := pagination.NewCursorPage([]string{}, []string{"2024-01-01T00:00:00Z", "my-id"}, nil)
		decoded, err := pagination.DecodeCursor(page.NextCursor)
		require.NoError(t, err)
		assert.Equal(t, []string{"2024-01-01T00:00:00Z", "my-id"}, decoded)
	})
}
