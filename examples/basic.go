package main

import (
	"flag"
	"fmt"
	"github.com/superlocrian/cfg"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	IntVal   int    `yaml:"intval" cmd:"intval|integer value" env:"INTVAL"`
	Int64    int64  `yaml:"int64"  cmd:"int64|integer 64 value " env:"INT64"`
	StrVal   string `yaml:"strval" cmd:"strval" env:"STRVAL"`
	IntSlice []int  `yaml:"int-slice" cmd:"intslice" env:"INTSLICE"`
	Data     struct {
		Fl64    float64 `yaml:"fl-64" cmd:"float|how to use float" env:"FL64"`
		BoolVal bool    `yaml:"bool-val" cmd:"bool|boolean value" env:"BOOL"`
	}
}

func (o *Config) FromYamlFile(file string) (err error) {
	var yml []byte
	yml, err = ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return yaml.Unmarshal(yml, o)
}

func main() {

	conf := &Config{}
	if err := conf.FromYamlFile("config.yml"); err != nil {
		log.Fatal(err)
	}
	if err := cfg.Unmarshal(conf, "", flag.CommandLine, os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", conf)

}
