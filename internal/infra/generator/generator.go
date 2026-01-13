package generator

type Generator interface {
	ShortURLKey() string
}

// Provide returns a Generator based on the config type.
// Currently only "simple" is supported. Add new generator types here as needed.
func Provide(cfg Config) Generator {
	switch cfg.Type {
	case "simple":
		return new(Simple)
	default:
		// Fall back to simple generator for unknown types
		return new(Simple)
	}
}
