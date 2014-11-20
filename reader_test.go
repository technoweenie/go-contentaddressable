package contentaddressable

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyOpen(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	if err := ioutil.WriteFile(filename, []byte("SUP"), 0755); err != nil {
		t.Fatal(err.Error())
	}

	reader, err := Open(filename)
	if err != nil {
		t.Fatal(err.Error())
	}

	by, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err.Error())
	}

	if content := string(by); content != "SUP" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestVerifyReader(t *testing.T) {
	buf := newBuffer("WAT")
	reader := Reader(buf, "d3f2dfc28bb4cbc063fb284734c102a38f96e41fa137dd77478015680fffd81e")

	by, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err.Error())
	}

	if content := string(by); content != "WAT" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestReadBadData(t *testing.T) {
	buf := newBuffer("WAT")
	reader := Reader(buf, "BAD-OID")

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
