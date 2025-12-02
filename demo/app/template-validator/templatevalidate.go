package main

import (
	"flag"
	"html/template"
)

var templateFile = flag.String("template", "/conf/index.tmpl", "Path to the template file to validate")

func main() {

	flag.Parse()

	template.Must(template.ParseFiles(*templateFile))
	// This is a placeholder for the template validator main function.
}
