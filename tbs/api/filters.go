package api

import (
	"fmt"
	"strings"
)

func NewFilterVal() *filter {
	return &filter{changed: false, value: make(map[string]string)}
}

type filter struct {
	value   map[string]string
	changed bool
}

func (f *filter) String() string {
	out := make([]string, len(f.value))
	i := 0
	for k, v := range f.value {
		out[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return "[" + strings.Join(out, " ") + "]"
}

func (f *filter) Set(val string) error {
	ss := strings.Split(val, ",")
	out := make(map[string]string, len(ss))
	for _, d := range ss {
		f := strings.Split(d, "=")
		out[f[0]] = f[1]
	}
	if !f.changed {
		f.value = out
	} else {
		new := make(map[string]string)
		for k, v := range f.value {
			new[k] = v
		}
		for k, v := range out {
			new[k] = v
		}
		f.value = new
	}
	f.changed = true
	return nil
}

func (f *filter) Type() string {
	return "filter"
}

func (f *filter) Get(key string) *string {
	val := f.value[key]
	return &val
}
