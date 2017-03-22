package api

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	format := ""
	target := &struct{ Test string }{}
	tableRenderer := NewRenderer(format, target)
	if _, ok := tableRenderer.(*TableRenderer); !ok {
		t.Error("Renderer without a format should be table renderer")
	}
	format = "json"
	target2 := &struct{ Test string }{}
	jsonRenderer := NewRenderer(format, target2)
	if _, ok := jsonRenderer.(*JSONRenderer); !ok {
		t.Error("Renderer with json format should be json renderer")
	}
}

func TestJSONRendererRender(t *testing.T) {
	target := &struct{ Test string }{"test"}
	var buf bytes.Buffer
	jr := &JSONRenderer{target: target}
	err := jr.Render(&buf)
	if err != nil {
		t.Error(err)
	}
	expected := `{
    "Test": "test"
}
`
	if buf.String() != expected {
		fmt.Println(buf.String(), expected)
		t.Error("Wrong json output")
	}
}

func TestTableRenderer(t *testing.T) {
	target := []struct {
		String string `json:"string"`
		Bool   bool   `json:"bool"`
		Int    int    `json:"int"`
	}{
		{
			String: "test",
			Bool:   false,
			Int:    1,
		},
	}
	format := "{{.String}}\t{{.Bool}}\t{{.Int}}"
	tr := &TableRenderer{format: format, target: target}
	var buf bytes.Buffer
	err := tr.Render(&buf)
	if err != nil {
		t.Error(err)
	}
	columns, err := buf.ReadString('\n')
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(columns, "String") {
		t.Error("No String column in output")
	}
	if !strings.Contains(columns, "Bool") {
		t.Error("No Bool column in output")
	}
	if !strings.Contains(columns, "Int") {
		t.Error("No Int column in output")
	}
	values, err := buf.ReadString('\n')
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(values, "test") {
		t.Error("No String value in output")
	}
	if !strings.Contains(values, "false") {
		t.Error("No Bool value in output")
	}
	if !strings.Contains(values, "1") {
		t.Error("No Int value in output")
	}
}
