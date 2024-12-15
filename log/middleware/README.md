<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# middleware

```go
import "github.com/brpaz/lib-go/log/middleware"
```

Package middleware provides a set of middleware functions for HTTP servers, related to logging, like a request logger.

## Index

- [func RequestLogger\(logger log.Logger, config \*RequestLoggerConfig\) func\(http.Handler\) http.Handler](<#RequestLogger>)
- [type LoggableResponseWriter](<#LoggableResponseWriter>)
  - [func \(lrw \*LoggableResponseWriter\) Write\(p \[\]byte\) \(int, error\)](<#LoggableResponseWriter.Write>)
  - [func \(lrw \*LoggableResponseWriter\) WriteHeader\(code int\)](<#LoggableResponseWriter.WriteHeader>)
- [type RequestLoggerConfig](<#RequestLoggerConfig>)
  - [func DefaultRequestLoggerConfig\(\) RequestLoggerConfig](<#DefaultRequestLoggerConfig>)


<a name="RequestLogger"></a>
## func [RequestLogger](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L89>)

```go
func RequestLogger(logger log.Logger, config *RequestLoggerConfig) func(http.Handler) http.Handler
```

RequestLogger is a middleware that logs the http request.

Usage:

```
package main
import (
	"net/http"
 	"github.com/brpaz/lib-go/log"
	"github.com/brpaz/lib-go/log/middleware"
)

func main() {
	logger, _ := log.New(log.WithAdapter("zap"))
	http.Handle("/", middleware.RequestLogger(logger, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	        w.Write([]byte("Hello, World!"))
	}))
	http.ListenAndServe(":8080", nil)
}
```

<a name="LoggableResponseWriter"></a>
## type [LoggableResponseWriter](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L13-L18>)

LoggableResponseWriter is a wrapper around http.ResponseWriter that captures the status code, so that we can retrieve it in the logging middleware.

```go
type LoggableResponseWriter struct {
    http.ResponseWriter
    StatusCode   int
    ResponseSize int
    Written      bool
}
```

<a name="LoggableResponseWriter.Write"></a>
### func \(\*LoggableResponseWriter\) [Write](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L27>)

```go
func (lrw *LoggableResponseWriter) Write(p []byte) (int, error)
```

Write writes the data to the response and tracks the size of the written data.

<a name="LoggableResponseWriter.WriteHeader"></a>
### func \(\*LoggableResponseWriter\) [WriteHeader](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L21>)

```go
func (lrw *LoggableResponseWriter) WriteHeader(code int)
```

WriteHeader writes the status code to the response.

<a name="RequestLoggerConfig"></a>
## type [RequestLoggerConfig](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L37-L50>)

RequestLoggerConfig holds configuration options for logging middleware.

```go
type RequestLoggerConfig struct {
    LogMethod             bool
    LogPath               bool
    LogStatus             bool
    LogDuration           bool
    LogRemoteAddr         bool
    LogUserAgent          bool
    LogRequestSize        bool
    LogResponseSize       bool
    LogRequestQueryParams bool
    LogRequestHeaders     bool
    LogRequestBody        bool
    LogResponseBody       bool
}
```

<a name="DefaultRequestLoggerConfig"></a>
### func [DefaultRequestLoggerConfig](<https://github.com/brpaz/lib-go/blob/main/log/middleware/request_logger.go#L67>)

```go
func DefaultRequestLoggerConfig() RequestLoggerConfig
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)