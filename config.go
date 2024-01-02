package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Env struct {
	Name    string                       `mapstructure:"env"`
	Tenants map[string]map[string]string `mapstructure:"tenants"`
}

type Script struct {
	Name         string `mapstructure:"name"`
	Command      string `mapstructure:"command"`
	Sequentially bool   `mapstructure:"sequentially"`
}

type Config struct {
	Envs       []Env                        `mapstructure:"envs"`
	Scripts    []Script                     `mapstructure:"scripts"`
	Variables  map[string]string            `mapstructure:"variables"`
	Postscript string                       `mapstructure:"postscript"`
	Prescript  string                       `mapstructure:"prescript"`
	Tenants    map[string]map[string]string `mapstructure:"tenants"`
}

func GetConfig() Config {

	vp := viper.New()
	// config := &Config{}

	vp.SetConfigName("ccrtly.config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("$HOME")
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
