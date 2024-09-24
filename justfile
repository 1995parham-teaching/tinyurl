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
    go run ariga.io/atlas/cmd/atlas@latest migrate diff --env local
    go run ariga.io/atlas/cmd/atlas@latest migrate apply --env local
    go run ./cmd/tinyurl/main.go seed

# connect into the dev environment database
database: (dev "up") (dev "exec" "database psql postgresql://tinyurl:secret@localhost/tinyurl")

# run golangci-lint linting
[group('lint')]
go-lint *flags:
    go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -c .golangci.yml {{ flags }}

# run atlas linting over migrations
[group('lint')]
atlas-lint:
    go run ariga.io/atlas/cmd/atlas@latest migrate lint --env local --git-base origin/main
