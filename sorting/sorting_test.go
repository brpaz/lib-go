package sorting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/sorting"
)

func TestParseDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    sorting.Direction
		wantErr bool
	}{
		{name: "asc uppercase", input: "ASC", want: sorting.ASC},
		{name: "desc uppercase", input: "DESC", want: sorting.DESC},
		{name: "asc lowercase", input: "asc", want: sorting.ASC},
		{name: "desc lowercase", input: "desc", want: sorting.DESC},
		{name: "mixed case", input: "Asc", want: sorting.ASC},
		{name: "invalid", input: "RANDOM", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sorting.ParseDirection(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSort(t *testing.T) {
	t.Parallel()

	allowed := []string{"created_at", "name", "email"}

	t.Run("valid field and direction", func(t *testing.T) {
		t.Parallel()

		s, err := sorting.NewSort("created_at", "DESC", allowed)
		require.NoError(t, err)
		assert.Equal(t, "created_at", s.Field)
		assert.Equal(t, sorting.DESC, s.Dir)
	})

	t.Run("field not in allow-list", func(t *testing.T) {
		t.Parallel()

		_, err := sorting.NewSort("password", "ASC", allowed)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"password"`)
	})

	t.Run("SQL injection attempt rejected", func(t *testing.T) {
		t.Parallel()

		_, err := sorting.NewSort("name; DROP TABLE users--", "ASC", allowed)
		require.Error(t, err)
	})

	t.Run("invalid direction", func(t *testing.T) {
		t.Parallel()

		_, err := sorting.NewSort("name", "SIDEWAYS", allowed)
		require.Error(t, err)
	})

	t.Run("empty allow-list rejects all fields", func(t *testing.T) {
		t.Parallel()

		_, err := sorting.NewSort("name", "ASC", []string{})
		require.Error(t, err)
	})
}

func TestSort_SQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		field string
		dir   sorting.Direction
		want  string
	}{
		{name: "asc", field: "created_at", dir: sorting.ASC, want: "created_at ASC"},
		{name: "desc", field: "name", dir: sorting.DESC, want: "name DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := sorting.NewSort(tt.field, string(tt.dir), []string{tt.field})
			require.NoError(t, err)
			assert.Equal(t, tt.want, s.SQL())
		})
	}
}

func TestSorts_SQL(t *testing.T) {
	t.Parallel()

	allowed := []string{"status", "created_at", "name"}

	t.Run("joins multiple clauses in order", func(t *testing.T) {
		t.Parallel()

		status, err := sorting.NewSort("status", "ASC", allowed)
		require.NoError(t, err)

		createdAt, err := sorting.NewSort("created_at", "DESC", allowed)
		require.NoError(t, err)

		sorts := sorting.Sorts{status, createdAt}
		assert.Equal(t, "status ASC, created_at DESC", sorts.SQL())
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		s, err := sorting.NewSort("name", "ASC", allowed)
		require.NoError(t, err)

		sorts := sorting.Sorts{s}
		assert.Equal(t, "name ASC", sorts.SQL())
	})

	t.Run("empty returns empty string", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "", sorting.Sorts{}.SQL())
		assert.Equal(t, "", sorting.Sorts(nil).SQL())
	})
}

func TestDefault(t *testing.T) {
	t.Parallel()

	t.Run("uses provided sort when field is set", func(t *testing.T) {
		t.Parallel()

		s, err := sorting.NewSort("name", "ASC", []string{"name", "created_at"})
		require.NoError(t, err)

		got := sorting.Default(s, "created_at", sorting.DESC)
		assert.Equal(t, "name", got.Field)
		assert.Equal(t, sorting.ASC, got.Dir)
	})

	t.Run("falls back to default when field is empty", func(t *testing.T) {
		t.Parallel()

		got := sorting.Default(sorting.Sort{}, "created_at", sorting.DESC)
		assert.Equal(t, "created_at", got.Field)
		assert.Equal(t, sorting.DESC, got.Dir)
	})
}
