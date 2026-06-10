package nullable_test

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/nullable"
)

// textAndScanType implements [encoding.TextMarshaler], [encoding.TextUnmarshaler],
// [sql.Scanner] and [driver.Valuer], for exercising those code paths in
// [nullable.Value].
type textAndScanType struct {
	v string
}

func (t textAndScanType) MarshalText() ([]byte, error) {
	return []byte("custom:" + t.v), nil
}

func (t *textAndScanType) UnmarshalText(text []byte) error {
	const prefix = "custom:"
	s := string(text)
	if !strings.HasPrefix(s, prefix) {
		return fmt.Errorf("invalid textAndScanType text %q", s)
	}
	t.v = strings.TrimPrefix(s, prefix)
	return nil
}

func (t *textAndScanType) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("textAndScanType: cannot scan %T", src)
	}
	t.v = s
	return nil
}

func (t textAndScanType) Value() (driver.Value, error) {
	return t.v, nil
}

func TestOf(t *testing.T) {
	t.Parallel()

	n := nullable.Of("hello")
	assert.True(t, n.Valid)
	assert.Equal(t, "hello", n.Val)
}

func TestEmpty(t *testing.T) {
	t.Parallel()

	n := nullable.Empty[int]()
	assert.False(t, n.Valid)
	assert.Equal(t, 0, n.Val)
}

func TestFromPtr(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer becomes invalid", func(t *testing.T) {
		t.Parallel()

		n := nullable.FromPtr[string](nil)
		assert.False(t, n.Valid)
	})

	t.Run("non-nil pointer becomes valid", func(t *testing.T) {
		t.Parallel()

		v := "hello"
		n := nullable.FromPtr(&v)
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.Val)
	})
}

func TestToPtr(t *testing.T) {
	t.Parallel()

	t.Run("invalid becomes nil", func(t *testing.T) {
		t.Parallel()

		p := nullable.Empty[string]().ToPtr()
		assert.Nil(t, p)
	})

	t.Run("valid becomes pointer to value", func(t *testing.T) {
		t.Parallel()

		p := nullable.Of("hello").ToPtr()
		require.NotNil(t, p)
		assert.Equal(t, "hello", *p)
	})

	t.Run("round trips through FromPtr", func(t *testing.T) {
		t.Parallel()

		v := 42
		n := nullable.FromPtr(&v)
		p := n.ToPtr()

		require.NotNil(t, p)
		assert.Equal(t, v, *p)
		assert.NotSame(t, &v, p)
	})
}

func TestString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "<null>", nullable.Empty[string]().String())
	assert.Equal(t, "hello", nullable.Of("hello").String())
	assert.Equal(t, "42", nullable.Of(42).String())
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		n         nullable.Value[string]
		wantValue string
		wantOk    bool
	}{
		{
			name:      "valid",
			n:         nullable.Of("foo"),
			wantValue: "foo",
			wantOk:    true,
		},
		{
			name:      "invalid",
			n:         nullable.Empty[string](),
			wantValue: "",
			wantOk:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, ok := tc.n.Get()
			assert.Equal(t, tc.wantValue, v)
			assert.Equal(t, tc.wantOk, ok)
		})
	}
}

func TestValueOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        nullable.Value[int]
		fallback int
		want     int
	}{
		{
			name:     "valid returns value",
			n:        nullable.Of(42),
			fallback: 0,
			want:     42,
		},
		{
			name:     "invalid returns fallback",
			n:        nullable.Empty[int](),
			fallback: 99,
			want:     99,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.n.ValueOr(tc.fallback))
		})
	}
}

func TestIsNull(t *testing.T) {
	t.Parallel()

	assert.False(t, nullable.Of("x").IsNull())
	assert.True(t, nullable.Empty[string]().IsNull())
}

func TestSet(t *testing.T) {
	t.Parallel()

	var n nullable.Value[string]
	n.Set("bar")
	assert.True(t, n.Valid)
	assert.Equal(t, "bar", n.Val)
}

func TestClear(t *testing.T) {
	t.Parallel()

	n := nullable.Of(123)
	n.Clear()
	assert.False(t, n.Valid)
	assert.Equal(t, 0, n.Val)
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    nullable.Value[any]
		want string
	}{
		{
			name: "valid string",
			n:    nullable.Of[any]("hello"),
			want: `"hello"`,
		},
		{
			name: "valid int",
			n:    nullable.Of[any](42),
			want: `42`,
		},
		{
			name: "null",
			n:    nullable.Empty[any](),
			want: `null`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tc.n)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(data))
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid string", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[string]
		require.NoError(t, json.Unmarshal([]byte(`"world"`), &n))
		assert.True(t, n.Valid)
		assert.Equal(t, "world", n.Val)
	})

	t.Run("null becomes invalid", func(t *testing.T) {
		t.Parallel()
		n := nullable.Of("existing")
		require.NoError(t, json.Unmarshal([]byte(`null`), &n))
		assert.False(t, n.Valid)
	})

	t.Run("valid int", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		require.NoError(t, json.Unmarshal([]byte(`7`), &n))
		assert.True(t, n.Valid)
		assert.Equal(t, 7, n.Val)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		err := json.Unmarshal([]byte(`"not-an-int"`), &n)
		require.Error(t, err)
		assert.False(t, n.Valid)
	})
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	type wrapper struct {
		Name nullable.Value[string] `json:"name"`
		Age  nullable.Value[int]    `json:"age"`
	}

	original := wrapper{
		Name: nullable.Of("Alice"),
		Age:  nullable.Empty[int](),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded wrapper
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Age, decoded.Age)
}

func TestScan(t *testing.T) {
	t.Parallel()

	t.Run("nil src sets null", func(t *testing.T) {
		t.Parallel()
		n := nullable.Of("existing")
		require.NoError(t, n.Scan(nil))
		assert.True(t, n.IsNull())
	})

	t.Run("string src sets valid", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[string]
		require.NoError(t, n.Scan("hello"))
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.Val)
	})

	t.Run("int64 src sets valid", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int64]
		require.NoError(t, n.Scan(int64(42)))
		assert.True(t, n.Valid)
		assert.Equal(t, int64(42), n.Val)
	})

	t.Run("convertible type src sets valid", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		require.NoError(t, n.Scan(int64(42)))
		assert.True(t, n.Valid)
		assert.Equal(t, 42, n.Val)
	})

	t.Run("incompatible type returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		err := n.Scan("not-an-int")
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("sql.Scanner implementation", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[textAndScanType]
		require.NoError(t, n.Scan("hello"))
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.Val.v)
	})

	t.Run("sql.Scanner implementation returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[textAndScanType]
		err := n.Scan(42)
		require.Error(t, err)
		assert.False(t, n.Valid)
	})
}

func TestValue(t *testing.T) {
	t.Parallel()

	t.Run("empty returns nil", func(t *testing.T) {
		t.Parallel()
		n := nullable.Empty[string]()
		v, err := n.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("valid string returns value", func(t *testing.T) {
		t.Parallel()
		n := nullable.Of("world")
		v, err := n.Value()
		require.NoError(t, err)
		assert.Equal(t, "world", v)
	})

	t.Run("driver.Valuer implementation", func(t *testing.T) {
		t.Parallel()
		n := nullable.Of(textAndScanType{v: "z"})
		v, err := n.Value()
		require.NoError(t, err)
		assert.Equal(t, "z", v)
	})
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	t.Run("invalid returns empty", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Empty[string]().MarshalText()
		require.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of("hello").MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of(42).MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "42", string(data))
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of(uint(42)).MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "42", string(data))
	})

	t.Run("TextMarshaler implementation", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of(textAndScanType{v: "x"}).MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "custom:x", string(data))
	})

	t.Run("float", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of(3.14).MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "3.14", string(data))
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()
		data, err := nullable.Of(true).MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "true", string(data))
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		t.Parallel()
		_, err := nullable.Of(struct{ X int }{X: 1}).MarshalText()
		require.Error(t, err)
	})
}

func TestUnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("empty input marks invalid", func(t *testing.T) {
		t.Parallel()
		n := nullable.Of("existing")
		require.NoError(t, n.UnmarshalText([]byte{}))
		assert.False(t, n.Valid)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[string]
		require.NoError(t, n.UnmarshalText([]byte("hello")))
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.Val)
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		require.NoError(t, n.UnmarshalText([]byte("42")))
		assert.True(t, n.Valid)
		assert.Equal(t, 42, n.Val)
	})

	t.Run("invalid int returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[int]
		err := n.UnmarshalText([]byte("abc"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[uint]
		require.NoError(t, n.UnmarshalText([]byte("42")))
		assert.True(t, n.Valid)
		assert.Equal(t, uint(42), n.Val)
	})

	t.Run("invalid uint returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[uint]
		err := n.UnmarshalText([]byte("-1"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("float", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[float64]
		require.NoError(t, n.UnmarshalText([]byte("3.14")))
		assert.True(t, n.Valid)
		assert.InDelta(t, 3.14, n.Val, 0.0001)
	})

	t.Run("invalid float returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[float64]
		err := n.UnmarshalText([]byte("abc"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[bool]
		require.NoError(t, n.UnmarshalText([]byte("true")))
		assert.True(t, n.Valid)
		assert.True(t, n.Val)
	})

	t.Run("invalid bool returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[bool]
		err := n.UnmarshalText([]byte("not-a-bool"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("TextUnmarshaler implementation", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[textAndScanType]
		require.NoError(t, n.UnmarshalText([]byte("custom:y")))
		assert.True(t, n.Valid)
		assert.Equal(t, "y", n.Val.v)
	})

	t.Run("TextUnmarshaler implementation returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[textAndScanType]
		err := n.UnmarshalText([]byte("not-custom"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		t.Parallel()
		var n nullable.Value[struct{ X int }]
		err := n.UnmarshalText([]byte("x"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})

	t.Run("round trips through MarshalText", func(t *testing.T) {
		t.Parallel()
		original := nullable.Of(7)

		data, err := original.MarshalText()
		require.NoError(t, err)

		var decoded nullable.Value[int]
		require.NoError(t, decoded.UnmarshalText(data))
		assert.Equal(t, original, decoded)
	})
}

func TestEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    nullable.Value[int]
		b    nullable.Value[int]
		want bool
	}{
		{name: "both invalid are equal", a: nullable.Empty[int](), b: nullable.Empty[int](), want: true},
		{name: "same value valid are equal", a: nullable.Of(5), b: nullable.Of(5), want: true},
		{name: "different value valid are not equal", a: nullable.Of(5), b: nullable.Of(6), want: false},
		{name: "valid and invalid are not equal", a: nullable.Of(5), b: nullable.Empty[int](), want: false},
		{name: "invalid and valid are not equal", a: nullable.Empty[int](), b: nullable.Of(5), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, nullable.Equal(tc.a, tc.b))
		})
	}
}
