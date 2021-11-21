package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/winebarrel/yktr2"
)

var version string

const (
	DefaultConfig = "yktr2.toml"
)

func parseArgs() *yktr2.Config {
	var config string
	flag.StringVar(&config, "config", "", "config file")
	ver := flag.Bool("version", false, "print version")
	flag.Parse()

	if *ver {
		printVersionAndEixt()
	}

	if config == "" {
		exePath, err := os.Executable()

		if err != nil {
			log.Fatal(err)
		}

		config = path.Join(filepath.Dir(exePath), DefaultConfig)
	}

	return loadConfig(config)
}

func loadConfig(path string) *yktr2.Config {
	rawCfg, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	yktr2Cfg := &yktr2.Config{}
	err = toml.Unmarshal(rawCfg, yktr2Cfg)

	if err != nil {
		log.Fatal(err)
	}

	err = yktr2Cfg.Validate()

	if err != nil {
		log.Fatal(err)
	}

	return yktr2Cfg
}

func printVersionAndEixt() {
	fmt.Fprintln(os.Stderr, version)
	os.Exit(0)
}
