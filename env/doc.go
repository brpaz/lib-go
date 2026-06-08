// Package env provides helpers for reading typed values from environment variables.
//
// Every Get* function accepts a key and a defaultValue. The default is returned
// when the variable is unset, empty, or (for numeric/duration helpers) unparseable
// — invalid values are never fatal.
//
// Every MustGet* function accepts only a key and panics when the variable is
// unset, empty, or (for numeric/duration helpers) unparseable. Use these for
// required configuration that has no sensible default (e.g. a database URL or
// API secret) and should fail fast at startup rather than run with a zero value.
//
// # Available helpers
//
//   - [GetString] / [MustGetString] — returns the raw string value
//   - [GetBool] / [MustGetBool] — truthy values are "true"/"1", falsy are "false"/"0"
//   - [GetInt] / [MustGetInt] — parses a decimal integer
//   - [GetFloat] / [MustGetFloat] — parses a 64-bit float
//   - [GetDuration] / [MustGetDuration] — parses a [time.Duration] string (e.g. "5m", "1h30m")
//   - [GetStringSlice] / [MustGetStringSlice] — splits a comma-separated value;
//     trims whitespace, drops empty elements
//
// # Example
//
//	port     := env.GetInt("PORT", 8080)
//	debug    := env.GetBool("DEBUG", false)
//	timeout  := env.GetDuration("REQUEST_TIMEOUT", 30*time.Second)
//	origins  := env.GetStringSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"})
//	dbURL    := env.MustGetString("DATABASE_URL")
package env
