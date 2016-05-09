package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/raboof/connbeat/beater"
)

var Name = "connbeat"

func main() {
	if err := beat.Run(Name, "", beater.New()); err != nil {
		os.Exit(1)
	}
}
