# Tiny URL

## Introduction

Writing an API project in Golang is Art, because there isn't any de-faco framework or standards to do something.
You have a planty of ways for doing things and at end you may figure out the way is not good or extensible as you think.

I want in this project select some of these ways and shows how their end is looking:

- Logging with [zap](https://github.com/uber-go/zap)
- Metrics with [otel](https://github.com/open-telemetry/opentelemetry-go)
- Tracing with [otel](https://github.com/open-telemetry/opentelemetry-go)

## Packing

I am following the rules defined by [golang-standard](https://github.com/golang-standards/project-layout).
The `internal/domain` package contains the domain-specific logics. As rule of thumbs everything defined in
`internal/domain` must use only go standard packages or other application packages, so they should use any third party
libraries directly.
