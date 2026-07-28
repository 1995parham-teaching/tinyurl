package generator

type Config struct {
	// Type selects the generation strategy, one of the Type constants.
	Type string `json:"type,omitempty" koanf:"type"`

	// Key is the secret the feistel generator permutes identifiers with, and is ignored by
	// every other type. It has no default: a shared one would make the permutation public and
	// hand anybody the ability to predict the next key.
	Key string `json:"key,omitempty" koanf:"key"`
}
