package contentaddressable

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

func Open(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return file, err
	}

	return Reader(file, filepath.Base(filename)), nil
}

func Reader(reader io.ReadCloser, oid string) io.ReadCloser {
	return &verifyingReader{oid, sha256.New(), reader}
}

type verifyingReader struct {
	ExpectedOid string
	Hash        hash.Hash
	io.ReadCloser
}

func (r *verifyingReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)

	if n > 0 {
		r.Hash.Write(p[0:n])
	}

	if err == io.EOF {
		oid := hex.EncodeToString(r.Hash.Sum(nil))
		if oid != r.ExpectedOid {
			return n, fmt.Errorf("Expected OID: %s, got: %s", r.ExpectedOid, oid)
		}
	}

	return n, err
}
