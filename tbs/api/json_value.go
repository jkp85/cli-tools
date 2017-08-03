package api

import (
	"bytes"
	"encoding/json"
	"log"
)

func NewJSONVal() *jsonValue {
	value := make(map[string]interface{})
	return &jsonValue{&value}
}

type jsonValue struct {
	Value interface{}
}

func (j *jsonValue) String() string {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(j.Value)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func (j *jsonValue) Set(s string) error {
	if s == "" {
		j.Value = nil
		return nil
	} else {
		buf := bytes.NewBufferString(s)
		return json.NewDecoder(buf).Decode(j.Value)
	}
}

func (j jsonValue) Type() string {
	return "json"
}
