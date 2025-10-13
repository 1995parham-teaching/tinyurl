package config

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/1995parham-teaching/tinyurl/internal/infra/db"
	"github.com/1995parham-teaching/tinyurl/internal/infra/generator"
	"github.com/1995parham-teaching/tinyurl/internal/infra/logger"
	"github.com/1995parham-teaching/tinyurl/internal/infra/telemetry"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/tidwall/pretty"
	"go.uber.org/fx"
)

// Config holds all configurations.
type Config struct {
	fx.Out

	Logger    logger.Config    `json:"logger"    koanf:"logger"`
	Generator generator.Config `json:"generator" koanf:"generator"`
	Database  db.Config        `json:"database"  koanf:"database"`
	Telemetry telemetry.Config `json:"telemetry" koanf:"telemetry"`
}

// prefix indicates environment variables prefix.
const prefix = "tinyurl_"

// Provide reads configuration with koanf.
func Provide() Config {
	k := koanf.New(".")

	// load default configuration from default function
	err := k.Load(structs.Provider(Default(), "koanf"), nil)
	if err != nil {
		log.Fatalf("error loading default: %s", err)
	}

	// load configuration from file
	err = k.Load(file.Provider("config.toml"), toml.Parser())
	if err != nil {
		log.Printf("error loading config.toml: %s", err)
	}

	// load environment variables
	err = k.Load(
		env.Provider(prefix, ".", func(source string) string {
			base := strings.ToLower(strings.TrimPrefix(source, prefix))

			return strings.ReplaceAll(base, "__", ".")
		}),
		nil,
	)
	if err != nil {
		log.Printf("error loading environment variables: %s", err)
	}

	var instance Config

	err = k.Unmarshal("", &instance)
	if err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	indent, err := json.MarshalIndent(instance, "", "\t")
	if err != nil {
		panic(err)
	}

	indent = pretty.Color(indent, nil)

	log.Printf(`
================ Loaded Configuration ================
%s
======================================================
	`, string(indent))

	return instance
}
