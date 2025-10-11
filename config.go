package main

import (
	"reflect"
	"os"
	"time"
	"fmt"
)

// Disclaimer:
// This is complete overkill, but fun to explore

type Config struct {
	RefreshInterval time.Duration `env:"REFRESH_INTERVAL" default:"10s"`
	APIToken        string        `env:"HCLOUD_TOKEN"`
}

func NewConfig() (*Config, error) {
	config := Config{}

	_type := reflect.TypeOf(config)
	configValue := reflect.ValueOf(&config)

	for i := range reflect.TypeOf(config).NumField() {
		field := _type.Field(i)
		env, envOk := field.Tag.Lookup("env")
		if !envOk {
			return nil, fmt.Errorf("env tag not specified for field %s `env:\"ENV_NAME\"`", field.Name)
		}

		valStr, defaultOk := field.Tag.Lookup("default")
		if envVal, envValOk := os.LookupEnv(env); envValOk {
			valStr = envVal
		} else if !envValOk && !defaultOk {
			return nil, fmt.Errorf("environment variable %s not set", env)
		}

		value := configValue.Elem().Field(i)

		switch value.Type() {
		case reflect.TypeOf(""):
			value.SetString(valStr)
		case reflect.TypeOf((time.Duration)(0)):
			intVal, err := time.ParseDuration(valStr)
			if err != nil {
				return nil, err
			}
			value.Set(reflect.ValueOf(intVal))
		default:
			return nil, fmt.Errorf("type %s of field %s not supported", field.Type, field.Name)
		}
	}

	return &config, nil
}
