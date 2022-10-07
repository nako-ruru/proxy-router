package main

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"runtime"
)

func main() {
	yamlFile, err := ioutil.ReadFile("./application.yml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	check(&config)
	start(&config)

	runtime.Goexit()
}
