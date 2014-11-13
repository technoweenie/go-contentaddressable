package contentaddressable

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestVerifyReader(t *testing.T) {
	buf := newBuffer("WAT")
	reader, err := ConsistentReader(buf, "d3f2dfc28bb4cbc063fb284734c102a38f96e41fa137dd77478015680fffd81e", 3)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	by, err := ioutil.ReadAll(reader)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if content := string(by); content != "WAT" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestReadSmallerData(t *testing.T) {
	buf := newBuffer("WAT")
	reader, err := ConsistentReader(buf, "d8689b62711dced3f13e45048ffff01759c9f65cc55206b3e95c054826f7f596", 2)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	by, err := ioutil.ReadAll(reader)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if content := string(by); content != "WA" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestReadBiggerData(t *testing.T) {
	buf := newBuffer("WA")
	reader, err := ConsistentReader(buf, "d3f2dfc28bb4cbc063fb284734c102a38f96e41fa137dd77478015680fffd81e", 3)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	by, err := ioutil.ReadAll(reader)

	if err == nil {
		t.Error("Expected error!")
	}

	if !strings.HasPrefix(err.Error(), "Expected OID:") {
		t.Errorf("Unexpected error: %s", err)
	}

	if content := string(by); content != "WA" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestReadBadData(t *testing.T) {
	buf := newBuffer("WAT")
	reader, err := ConsistentReader(buf, "BAD-OID", 3)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	by, err := ioutil.ReadAll(reader)

	if err == nil {
		t.Error("Expected error!")
	}

	if !strings.HasPrefix(err.Error(), "Expected OID:") {
		t.Errorf("Unexpected error: %s", err)
	}

	if content := string(by); content != "WAT" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func newBuffer(contents string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(contents))
}
