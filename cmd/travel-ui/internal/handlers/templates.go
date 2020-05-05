package handlers

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
)

var views = make(map[string]*template.Template)

func loadTemplate(name string, path string) error {
	pwd, _ := os.Getwd()
	path = pwd + "/views/" + path

	if _, exists := views[name]; exists {
		return errors.New("template already loaded")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "reading template")
	}

	tmpl, err := template.New(name).Parse(string(data))
	if err != nil {
		return errors.Wrap(err, "parsing template")
	}

	views[name] = tmpl
	return nil
}

func executeTemplate(name string, vars map[string]interface{}) []byte {
	var markup bytes.Buffer
	if err := views[name].Execute(&markup, vars); err != nil {
		log.Println(err)
		return []byte("error processing template")
	}
	return markup.Bytes()
}
