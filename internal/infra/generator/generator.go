package generator

type Generator interface {
	ShortURLKey() string
}

func Provide(cfg Config) Generator {
	if cfg.Type == "simple" {
		return new(Simple)
	}

	return new(Simple)
}
