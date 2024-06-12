package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

func LoadConfig(dest any) {
	if err := cleanenv.ReadEnv(dest); err != nil {
		log.Fatalf("failed to load env config: %s", err)
	}
}
