// Package nullable provides a generic [Value] type that represents a value
// that may or may not be present — like [sql.NullString] but for any type T.
//
// # Constructing
//
// Use [Of] to wrap a present value and [Empty] to represent absence:
//
//	name := nullable.Of("alice")      // valid, Val = "alice"
//	none := nullable.Empty[string]()  // invalid (null)
//
// [FromPtr] converts from pointer-based APIs (generated DTOs, ORMs) where
// optional fields are represented as *T:
//
//	var p *string
//	n := nullable.FromPtr(p)  // invalid — p is nil
//	p  = ptr.Of("alice")
//	n  = nullable.FromPtr(p)  // valid, Val = "alice"
//
// # Reading
//
// [Value.Get] returns the value and a boolean, similar to a map lookup:
//
//	if v, ok := name.Get(); ok {
//	    fmt.Println(v) // "alice"
//	}
//
// [Value.ValueOr] is a one-liner with a fallback:
//
//	display := name.ValueOr("anonymous")
//
// [Value.ToPtr] converts back to a pointer — nil when invalid — for
// round-tripping through pointer-based APIs:
//
//	p := name.ToPtr()  // *string pointing to "alice"
//
// [Value.String] implements [fmt.Stringer] for logging and debugging:
//
//	fmt.Sprint(name)             // "alice"
//	fmt.Sprint(nullable.Empty[string]()) // "<null>"
//
// # Mutating
//
//	n.Set("bob")  // marks valid, sets Val
//	n.Clear()     // marks invalid, zeroes Val
//
// # JSON
//
// Value marshals to the underlying JSON value when valid, or JSON null when not:
//
//	json.Marshal(nullable.Of("alice"))    // → "alice"
//	json.Marshal(nullable.Empty[string]()) // → null
//
//	var n nullable.Value[string]
//	json.Unmarshal([]byte(`"bob"`), &n) // n.Valid=true,  n.Val="bob"
//	json.Unmarshal([]byte(`null`),  &n) // n.Valid=false
//
// # Text encoding
//
// Value implements [encoding.TextMarshaler] and [encoding.TextUnmarshaler] —
// an empty string round-trips as invalid (null), and any other value is
// formatted/parsed as text. This is what JSON map keys, [url.Values] and
// form/query encoders use for non-string and optional values. T must be a
// string, a basic numeric/bool kind, or implement the encoding.Text*
// interfaces itself; any other type returns an error:
//
//	nullable.Of(42).MarshalText()           // → []byte("42"), nil
//	nullable.Empty[int]().MarshalText()      // → []byte{}, nil
//
//	var n nullable.Value[int]
//	n.UnmarshalText([]byte("42")) // n.Valid=true, n.Val=42
//	n.UnmarshalText([]byte(""))   // n.Valid=false
//
// # SQL
//
// Value implements [sql.Scanner] and [driver.Valuer] for use with database/sql:
//
//	var bio nullable.Value[string]
//	row.Scan(&bio) // NULL → bio.Valid=false; "x" → bio.Valid=true, bio.Val="x"
//
//	db.Exec("UPDATE users SET bio = ?", nullable.Of("new bio"))
//	db.Exec("UPDATE users SET bio = ?", nullable.Empty[string]()) // sends NULL
//
// # Equality
//
// [Equal] compares two Values of a comparable T:
//
//	nullable.Equal(nullable.Of(5), nullable.Of(5))         // true
//	nullable.Equal(nullable.Empty[int](), nullable.Empty[int]()) // true — both invalid
//	nullable.Equal(nullable.Of(5), nullable.Empty[int]())  // false
package nullable
