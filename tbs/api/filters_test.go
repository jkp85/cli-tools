package api

import "testing"

func TestNewFilterVal(t *testing.T) {
	var i interface{}
	var f *filter
	var ok bool
	i = NewFilterVal()
	if f, ok = i.(*filter); !ok {
		t.Error("NewFilterVal returned wrong type")
	}
	if f.changed != false {
		t.Error("Changed should be false in new filter val")
	}
}

//func TestFilterString(t *testing.T) {
//	f := NewFilterVal()
//	vals := "test=1,test2=test"
//	f.Set(vals)
//	expectedString := "[test=1 test2=test]"
//	if f.String() != expectedString {
//		t.Errorf("Wrong output: %s | %s", f, expectedString)
//	}
//}

func TestFilterSet(t *testing.T) {
	f := NewFilterVal()
	vals := "test=1,test2=test"
	err := f.Set(vals)
	if err != nil {
		t.Error(err)
	}
	if !f.changed {
		t.Error("Changed should be true after setting a value")
	}
	if f.value["test"] != "1" {
		t.Error("Wrong value for test")
	}
	if f.value["test2"] != "test" {
		t.Error("Wrong value for test2")
	}
}

func TestFilterGet(t *testing.T) {
	f := NewFilterVal()
	vals := "test=1"
	f.Set(vals)
	result := f.Get("test")
	if *result != "1" {
		t.Error("Wrong test value")
	}
}
