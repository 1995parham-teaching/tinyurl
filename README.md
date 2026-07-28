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
- **Migrations using [`goose`](https://github.com/pressly/goose):** Managing database schema migrations with embedded SQL files.

## Packaging

I am following the rules defined by [golang-standard](https://github.com/golang-standards/project-layout).
The `internal/domain` package contains the domain-specific logics. As rule of thumbs everything defined in
`internal/domain` must use only go standard packages or other application packages, so they should not use any third party
libraries directly.

The infrastructure layer does the actual using of third party libraries and resides in `infra` package.
Actual implementation always goes into the `infra` package.

## Repositories

To facilitate database access, we've defined repository interfaces within the `domain` package.
These form the contract for our data access methods.
Corresponding implementations can be found in the `infra` package,
ensuring a separation of concerns between our domain definitions and infrastructure-specific code.

### Repository Interfaces

- The interfaces are prefixed with `repo` to denote their role as repositories within the domain layer.

### Implementation

- The actual implementations carry a `db` prefix, indicating their direct interaction with the database
  and their role within the infrastructure layer.

## Key Generation

Handing out a short key that nobody else holds is the one problem a url shortener cannot avoid,
and there is more than one way to solve it. The `generator` package implements four, behind a single
interface, so they can be compared by changing `generator.type` in the configuration.

| Type      | Where uniqueness comes from                | Keys are     |
| --------- | ------------------------------------------ | ------------ |
| `simple`  | the primary key on `urls`, plus a retry    | random, guessable |
| `secure`  | the primary key on `urls`, plus a retry    | random, unguessable |
| `counter` | a database sequence, by construction       | sequential, enumerable |
| `feistel` | a database sequence, by construction       | scattered, unguessable |

The random generators make no promise of their own. They draw a key, the primary key on `urls` rejects it
if it is taken, and the service draws another — up to `urlsvc.MaxKeyGenAttempts` times. This holds up
because the key space stays sparse: with N keys stored, a fresh one collides with probability N/`Space`,
and the expected number of attempts is 1/(1-load factor). `simple` draws from `math/rand` and `secure`
from `crypto/rand`; only the latter produces keys a stranger cannot guess.

The counting generators never collide at all. `counter` encodes a Postgres sequence in base62, so distinct
identifiers give distinct keys and there is nothing to check and nothing to retry. Its weakness is that
consecutive keys are adjacent, so holding one lets anybody walk the rest.

`feistel` fixes that without giving up the guarantee. It pushes the identifier through a keyed
[Feistel network](https://en.wikipedia.org/wiki/Feistel_cipher) before encoding it. A Feistel network is a
bijection whatever its round function computes, so distinct identifiers still give distinct keys, while
consecutive ones land far apart. It requires `generator.key` to be set; that secret is what makes the
permutation unguessable, and it deliberately has no default.

The sequence behind both counting generators is declared with a cache, so each database session claims a
block of identifiers up front and spends none of them on coordination — the same idea as a hand written
Hi/Lo allocator. The cost is gaps in the numbering, which cost nothing but key space.

### One Key Space

Names chosen by a caller and keys produced by a generator live in the same key space, stored exactly as
the caller will type them. Nothing keeps the two apart because nothing needs to: the primary key rejects a
name that is taken, and a generated key that happens to land on a claimed name is simply regenerated.

Chosen names used to be stored behind a `static_` prefix instead. That guaranteed separation, but it meant
the key that was stored was not the key anybody typed, so every lookup of a chosen name spent one query
failing to find the typed key before trying the prefixed one. Uniqueness was already guaranteed elsewhere,
so the prefix bought a second query and nothing else.

## The Read Path

