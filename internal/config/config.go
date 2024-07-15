package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
)

type environment struct {
	PriceProviderURI  string `env:"PRICE_PROVIDER_URI" envDefault:"localhost:43223" validate:"uri"`
	PostgresURI       string `env:"POSTGRES_DB_URI" envDefault:"postgres://user:password@localhost:40333/chart?sslmode=disable" validate:"uri"`
	PositionServerURI string `env:"POSITION_SERVER_URI" envDefault:"localhost:7074" validate:"uri"`
	BalanceServerURI  string `env:"BALANCE_SERVER_URI" envDefault:"localhost:7075" validate:"uri"`
}

func New() (conf environment, err error) {
	if err = env.Parse(&conf); err != nil {
		return environment{}, err
	}

	val := validator.New(validator.WithRequiredStructEnabled())
	if err = val.Struct(&conf); err != nil {
		return environment{}, err
	}

	return conf, nil
}
