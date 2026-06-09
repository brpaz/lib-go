// Package httpapi provides helpers for building JSON HTTP APIs.
//
// It covers request decoding, structured error responses, and response writing.
// Error handling maps [errs.Error] domain errors to HTTP status codes and
// structured JSON bodies without leaking internal details to clients.
//
// # Request decoding
//
//	body, err := httpapi.DecodeJSON[CreateUserRequest](r)
//	if err != nil {
//		httpapi.HandleErr(w, r, err, nil)
//		return
//	}
//
// # Error handling
//
// [HandleErr] maps any error to a JSON response. It understands [errs.Error]
// values extracted from the chain, JSON decode errors, and falls back to 500
// for unknown errors. App-specific error codes are supported via extraCodes:
//
//	var appCodes = map[string]int{
//		"payment_failed": http.StatusPaymentRequired,
//	}
//
//	httpapi.HandleErr(w, r, err, appCodes)
//
// For [validator.ValidationError] from request validation, use [HandleValidationError]
// to tag each field error with its request location (body, query, path, header):
//
//	httpapi.HandleValidationError(w, r, httpapi.LocationBody, ve)
//
// # Response writing
//
//	httpapi.Ok(w, data)         // 200
//	httpapi.Created(w, data)    // 201
//	httpapi.NoContent(w)        // 204
package httpapi
