package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

func main()  {
	filename, err := filepath.Abs("./application.yml")
	if err != nil {
		panic(err)
	}
	yamlFile, err := ioutil.ReadFile(filename)
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
