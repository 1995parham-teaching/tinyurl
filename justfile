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
      docker compose -f ./deployments/docker-compose.yml up -d {{ flags }}
    else
      docker compose -f ./deployments/docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run tests in the dev environment
test $tinyurl_telemetry__meter__enabled="false": (dev "up")
    just seed
    go test -v ./... -covermode=atomic -coverprofile=coverage.out

seed $tinyurl_telemetry__meter__enabled="false": (dev "up")
    go run ./cmd/tinyurl/main.go migrate
    go run ./cmd/tinyurl/main.go seed

# connect into the dev environment database
database: (dev "up") (dev "exec" "database psql postgresql://tinyurl:secret@localhost/tinyurl")

# run golangci-lint
lint *flags:
    golangci-lint run -c .golangci.yml {{ flags }}
