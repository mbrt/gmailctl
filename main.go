package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func readConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var res Config
	err = yaml.Unmarshal(b, &res)
	return res, err

}

func main() {
	cfg, err := readConfig("example.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cfg)
}
