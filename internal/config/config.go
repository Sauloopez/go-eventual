package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

type RabbitMQConfig struct {
	User     string `env:"RABBITMQ_USER,required"`
	Password string `env:"RABBITMQ_PASSWORD,required"`
	Host     string `env:"RABBITMQ_HOST,required"`
	Port     string `env:"RABBITMQ_PORT,required"`
	Vhost    string `env:"RABBITMQ_VHOST"`
	Queue    string `env:"RABBITMQ_QUEUE,required"`
	MaxDelay string `env:"RABBITMQ_MAX_DELAY"`
}

type Config struct {
	DBPath         string `env:"DB_PATH,required"`
	RabbitMQConfig *RabbitMQConfig
}

func validateStructFields[T any](config *T) (*T, error) {
	v := reflect.ValueOf(config).Elem()
	t := reflect.TypeOf(config).Elem()
	logger := log.Default()

	logger.Printf("Validating struct fields for type: %s", t.Name())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		envVar := tagParts[0]
		logger.Printf("Loading variable '%s'", envVar)
		isRequired := len(tagParts) > 1 && tagParts[1] == "required"

		// Validar campos obligatorios
		envValue := os.Getenv(envVar)

		if isRequired && envValue == "" {
			return nil, fmt.Errorf("Environment variable '%s' is required but not configured", envVar)
		}

		// Asignar valor al campo
		if value.CanSet() {
			value.Set(reflect.ValueOf(envValue))
		} else {
			return nil, fmt.Errorf("Cannot set value for field '%s'", field.Name)
		}
	}

	return config, nil
}

func BuildConfig() (*Config, error) {
	rabbitConfig, err := validateStructFields(&RabbitMQConfig{})
	if err != nil {
		return nil, err
	}
	config, err := validateStructFields(&Config{})
	if err != nil {
		return nil, err
	}
	config.RabbitMQConfig = rabbitConfig
	return config, nil
}
