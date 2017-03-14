package contentaddressable

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var supOid = "a2b71d6ee8997eb87b25ab42d566c44f6a32871752c7c73eb5578cb1182f7be0"

func TestFileAccept(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	// init file
	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, filename, aw.filename)

	// write to file
	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	// check write is saved to temp file
	by, err := ioutil.ReadFile(aw.tempFilename)
	assertEqual(t, nil, err)
	assertEqual(t, "SUP", string(by))

	// check nothing saved to actual file yet
	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, 0, len(by))

	// file is accepted properly
	assertEqual(t, nil, aw.Accept())

	// file has final contents
	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, "SUP", string(by))

	// tempfile is gone
	by, err = ioutil.ReadFile(aw.tempFilename)
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not exist error, got: %+v", err)
	}
	assertEqual(t, 0, len(by))

	// no problem closing the file
	assertEqual(t, nil, aw.Close())
}

type badFile struct {
	CloseErr error
	SyncErr  error
	*os.File
}

func (f *badFile) Close() error {
	if f.CloseErr != nil {
		return f.CloseErr
	}
	return f.File.Close()
}

func (f *badFile) Sync() error {
	if f.SyncErr != nil {
		return f.SyncErr
	}
	return f.File.Sync()
}

func TestFileBadAcceptFileClose(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	// init file
	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, filename, aw.filename)

	badDestFile := &badFile{File: aw.file.(*os.File)}
	aw.file = badDestFile

	// write to file
	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	// check write is saved to temp file
	by, err := ioutil.ReadFile(aw.tempFilename)
	assertEqual(t, nil, err)
	assertEqual(t, "SUP", string(by))

	// check nothing saved to actual file yet
	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, 0, len(by))

	badDestFile.CloseErr = errors.New("test error")

	err = aw.Accept()
	if err != nil {
		assertEqual(t, "test error", err.Error())
	} else {
		t.Error("Accept should return error")
	}

	// file is gone
	by, err = ioutil.ReadFile(aw.filename)
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not exist error, got: %+v", err)
	}
	assertEqual(t, 0, len(by))

	// tempfile is gone
	by, err = ioutil.ReadFile(aw.tempFilename)
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not exist error, got: %+v", err)
	}
	assertEqual(t, 0, len(by))
}

func TestFileMismatch(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, "b2b71d6ee8997eb87b25ab42d566c44f6a32871752c7c73eb5578cb1182f7be0")
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)

	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	by, err := ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, 0, len(by))

	err = aw.Accept()
	if err == nil || !strings.Contains(err.Error(), "Content mismatch") {
		t.Errorf("Expected mismatch error: %s", err)
	}

	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, "", string(by))

	assertEqual(t, nil, aw.Close())

	_, err = ioutil.ReadFile(filename)
	assertEqual(t, true, os.IsNotExist(err))
}

func TestFileCancel(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)

	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	assertEqual(t, nil, aw.Close())

	for _, name := range []string{aw.filename, aw.tempFilename} {
		if _, err := os.Stat(name); err == nil {
			t.Errorf("%s exists?", name)
		}
	}
}

func TestFileLocks(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewWithSuffix(filename, "-wat")
	assertEqual(t, nil, err)
	assertEqual(t, filename, aw.filename)
	assertEqual(t, filename+"-wat", aw.tempFilename)

	files := []string{aw.filename, aw.tempFilename}

	for _, name := range files {
		if _, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0665); err == nil {
			t.Errorf("Able to open %s!", name)
		}
	}

	assertEqual(t, nil, aw.Close())

	for _, name := range files {
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0665)
		assertEqualf(t, nil, err, "unable to open %s: %s", name, err)
		cleanupFile(f, name)
	}
}

func TestFileDuel(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)
	defer aw.Close()

	if _, err := NewFile(filename); err == nil {
		t.Errorf("Expected a file open conflict!")
	}
}

func SetupFile(t *testing.T) *FileTest {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err.Error())
	}

	path := filepath.Join(wd, "File")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err.Error())
	}

	return &FileTest{path, t}
}

type FileTest struct {
	Path string
	*testing.T
}

func (t *FileTest) Teardown() {
	if err := os.RemoveAll(t.Path); err != nil {
		t.Fatalf("Error removing %s: %s", t.Path, err)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	checkAssertion(t, expected, actual, "")
}

func assertEqualf(t *testing.T, expected, actual interface{}, format string, args ...interface{}) {
	checkAssertion(t, expected, actual, format, args...)
}

func checkAssertion(t *testing.T, expected, actual interface{}, format string, args ...interface{}) {
	if expected == nil {
		if actual == nil {
			return
		}
	} else if reflect.DeepEqual(expected, actual) {
		return
	}

	_, file, line, _ := runtime.Caller(2) // assertEqual + checkAssertion
	t.Logf("%s:%d\nExpected: %v\nActual:   %v", file, line, expected, actual)
	if len(args) > 0 {
		t.Logf("! - "+format, args...)
	}
	t.FailNow()
}
