package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/viper"
)

func Render(formatName string, target interface{}) error {
	format := viper.GetString(formatName)
	renderer := NewRenderer(format, target)
	return renderer.Render(os.Stdout)
}

func NewRenderer(format string, target interface{}) Renderer {
	switch strings.SplitN(format, " ", 2)[0] {
	case "json":
		return &JSONRenderer{target}
	default:
		return &TableRenderer{
			target: target,
			format: format,
		}
	}
}

type Renderer interface {
	Render(io.Writer) error
}

type JSONRenderer struct {
	target interface{}
}

func (j *JSONRenderer) Render(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(j.target)
}

type TableRenderer struct {
	target interface{}
	format string
}

func (t *TableRenderer) Render(w io.Writer) error {
	format := t.processFormat()
	tmpl, err := template.New("table").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, t.target)
}

func (t *TableRenderer) processFormat() string {
	replacer := strings.NewReplacer("{{", "", "}}", "", ".", "", " ", "")
	columns := replacer.Replace(t.format)
	return fmt.Sprintf("%s\n{{ range . }}%s\n{{ end }}", columns, t.format)
}
