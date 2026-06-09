// Package openapitest provides utilities for validating HTTP requests and
// responses against an OpenAPI spec in tests.
//
// Typical usage with a package-level validator initialized in TestMain:
//
//	var v *openapitest.Validator
//
//	func TestMain(m *testing.M) {
//		var err error
//		v, err = openapitest.NewValidator("testdata/openapi.yaml")
//		if err != nil {
//			log.Fatal(err)
//		}
//		os.Exit(m.Run())
//	}
//
//	func TestSomething(t *testing.T) {
//		// ... make request ...
//		v.AssertResponse(t, req, resp)
//	}
package openapitest
