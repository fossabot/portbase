package record

import (
	"bytes"
	"testing"

	"github.com/Safing/portbase/formats/dsd"
)

func TestWrapper(t *testing.T) {

	// check model interface compliance
	var m Record
	w := &Wrapper{}
	m = w
	_ = m

	// create test data
	testData := []byte(`J{"a": "b"}`)

	// test wrapper
	wrapper, err := NewWrapper("test:a", &Meta{}, testData)
	if err != nil {
		t.Fatal(err)
	}
	if wrapper.Format != dsd.JSON {
		t.Error("format mismatch")
	}
	if !bytes.Equal(testData, wrapper.Data) {
		t.Error("data mismatch")
	}

	encoded, err := wrapper.Marshal(dsd.JSON)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(testData, encoded) {
		t.Error("marshal mismatch")
	}

	wrapper.SetMeta(&Meta{})
	raw, err := wrapper.MarshalRecord()
	if err != nil {
		t.Fatal(err)
	}

	wrapper2, err := NewRawWrapper("test", "a", raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(testData, wrapper2.Data) {
		t.Error("marshal mismatch")
	}

}
