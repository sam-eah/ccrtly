package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Env struct {
	Name    string                       `mapstructure:"env"`
	Tenants map[string]map[string]string `mapstructure:"tenants"`
}

type Config struct {
	Envs    []Env             `mapstructure:"envs"`
	Scripts map[string]string `mapstructure:"scripts"`
}

func GetConfig() Config {

	vp := viper.New()
	// config := &Config{}

	vp.SetConfigName("project")
	vp.SetConfigType("yaml")
	vp.AddConfigPath(".")

	err := vp.ReadInConfig()
	if err != nil {
		fmt.Println("Config not found...", err)
	}
	// fmt.Println(vp.ConfigFileUsed())

	config := &Config{}
	err = vp.Unmarshal(&config)
	if err != nil {
		fmt.Println(err)
	}

	proj := *config

	return proj
}
