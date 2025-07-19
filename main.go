package main

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
)

func main() {
	yamlFile, err := os.ReadFile("./application.yml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	for _, target := range config.Backends.Targets {
		target.Id = gonanoid.Must()
	}
	check(&config)
	start(&config)

	runtime.Goexit()
}
