beats-output-http
=================

Outputter for the Elastic Beats platform that simply
POSTs events to an HTTP endpoint.

[![Build Status](https://travis-ci.org/raboof/beats-output-http.svg?branch=master)](https://travis-ci.org/raboof/beats-output-http)

Usage
=====

To add support for this output plugin to a beat, you
have to import this plugin into your main beats package,
like this:

```
package main

import (
	"os"

	_ "github.com/raboof/beats-output-http"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/raboof/connbeat/beater"
)

var Name = "connbeat"

func main() {
	if err := beat.Run(Name, "", beater.New); err != nil {
		os.Exit(1)
	}
}
```

Then configure the http output plugin in yourbeat.yaml:

```
output:
  http:
    hosts: ["some.example.com:80/foo"]
```
