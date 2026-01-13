package server

type Config struct {
	Address string `json:"address,omitempty" koanf:"address"`
}
