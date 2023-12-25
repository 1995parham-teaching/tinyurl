<h1 align="center"> Tiny URL </h1>

<p align="center">
<img src="./.github/assets/logo.png" height="250px">
</p>

<p align="center">
    <img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham-teaching/tinyurl/test.yaml?logo=github&style=for-the-badge">
    <img alt="Codecov" src="https://img.shields.io/codecov/c/github/1995parham-teaching/tinyurl?logo=codecov&style=for-the-badge">
    <img alt="GitHub repo size" src="https://img.shields.io/github/repo-size/1995parham-teaching/tinyurl?logo=github&style=for-the-badge">
 </p>

## Introduction

Writing an API project in Golang can be likened to art,
due to the absence of a de facto framework or established standards for development.
There is a plethora of approaches available,
and it isn't uncommon to discover that the chosen method isn't as effective or extensible as initially assumed.

### Objectives

In this project, I aim to explore and demonstrate the outcomes of several such approaches:

- **Logging with [`zap`](https://github.com/uber-go/zap):** A fast, structured, leveled logging in Go.
- **Metrics with [`otel`](https://github.com/open-telemetry/opentelemetry-go) (OpenTelemetry):** Instrumenting code to collect and report metrics.
- **Tracing with [`otel`](https://github.com/open-telemetry/opentelemetry-go) (OpenTelemetry):** Capturing the flow and latency of operations in our application.
- **Dependency Injection using [`fx`](https://github.com/uber-go/fx):** A framework for dependency injection providing a robust way of managing dependencies.
- **Migrations using [`atlasgo`](https://atlasgo.io/):** Managing database schema migrations in an agile manner.

## Packaging

I am following the rules defined by [golang-standard](https://github.com/golang-standards/project-layout).
The `internal/domain` package contains the domain-specific logics. As rule of thumbs everything defined in
`internal/domain` must use only go standard packages or other application packages, so they should not use any third party
libraries directly.

The infrastructure layer does the actual using of third party libraries and resides in `infra` package.
Actual implementation always goes into the `infra` package.
