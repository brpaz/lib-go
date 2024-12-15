# lib-go

> A library of reusable code for my Go projects.

## Motivation

When creating Go projects, there are some code blocks and keep repeating in every project. Ex: Logging, glue code, utilities, etc.

The purpose of this repo, is to provide a personal library of that shared code, so I stop having to repeat it in all the my projects.

## What is included

- HTTP utilities to work with API responses
- HTTP Middlewares (Request Logger, Context, API Key validation)
- Logging
- Health checks
- Validation

### Dependencies

One of the great features of Go is it´s rich standard library. This library tries to use standard library functions (Ex: `net/http`) to make it compatible with different projects and reduce the number of external dependencies.

Still, this is an opinionated library and some abstractions requires external dependencies. (Ex: `zap` for logging).

## Getting started

### Install

```
go get github.com/brpaz/lib-go
```


