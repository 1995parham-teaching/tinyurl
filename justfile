default:
    @just --list

# build tinyurl binary
build:
    go build -o tinyurl ./cmd/tinyurl

# update go packages
update:
    @cd ./cmd/tinyurl && go get -u

# set up the dev environment with docker-compose
dev cmd *flags:
    #!/usr/bin/env bash
    set -euxo pipefail
    if [ {{ cmd }} = 'down' ]; then
      docker compose -f ./deployments/docker-compose.yml down --remove-orphans
      docker compose -f ./deployments/docker-compose.yml rm
    elif [ {{ cmd }} = 'up' ]; then
      docker compose -f ./deployments/docker-compose.yml up --wait -d {{ flags }}
    else
      docker compose -f ./deployments/docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run tests in the dev environment
test: (dev "up")
    just seed
    go test -v -race -shuffle=on ./... -covermode=atomic -coverprofile=coverage.out

seed: (dev "up")
    atlas migrate diff --env local
    atlas migrate apply --env local
    go run ./cmd/tinyurl/main.go seed

# connect into the dev environment database
database: (dev "up") (dev "exec" "database psql postgresql://tinyurl:secret@localhost/tinyurl")

# run golangci-lint linting
[group("lint")]
go-lint *flags:
    golangci-lint run -c .golangci.yml {{ flags }}

# generate a new migration file comparing the current state (from migrations dir/dev db) with the desired state defined by schema_source.
atlas-diff name="":
    @echo "==> Generating migration diff: {{ snakecase(name) }}..."
    atlas migrate diff {{ snakecase(name) }} --env local

# apply pending migrations to the database specified by db_url.
atlask-apply:
    @echo "==> Applying migrations to local environment ..."
    atlas migrate apply --env local

# show sql for pending migrations without applying them.
atlas-apply-dry:
    @echo "==> Dry-run applying migrations to local environment ..."
    atlas migrate apply --dry-run --env local

# lint migrations for potential issues (uses dev db). lints the latest migration by default.
[group("lint")]
atlas-lint N='1':
    @echo "==> Linting latest {{ N }} migration(s) using local environment ..."
    atlas migrate lint --latest {{ N }} --env local

# inspect the current schema of the live database (db_url).
inspect:
    @echo "==> Inspecting schema of local environment ..."
    atlas schema inspect --env local

# check if atlas cli is installed
check-atlas:
    @atlas version || (echo "Error: Atlas CLI not found. Install from https://atlasgo.io/cli/getting-started/setting-up" && exit 1)
