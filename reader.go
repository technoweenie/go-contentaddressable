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

func Open(filename string, expectedSize int64) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return file, err
	}

	return Reader(file, filepath.Base(filename), expectedSize), nil
}

func Reader(reader io.ReadCloser, oid string, size int64) io.ReadCloser {
	return &verifyingReader{oid, size, sha256.New(), reader}
}

type verifyingReader struct {
	ExpectedOid    string
	BytesRemaining int64
	Hash           hash.Hash
	io.ReadCloser
}

func (r *verifyingReader) Read(p []byte) (int, error) {
	if int64(len(p)) > r.BytesRemaining {
		p = p[0:r.BytesRemaining]
	}

	n, err := r.ReadCloser.Read(p)

	if n > 0 {
		r.Hash.Write(p[0:n])
		r.BytesRemaining -= int64(n)
	}

	if err == io.EOF || (err == nil && r.BytesRemaining <= 0) {
		oid := hex.EncodeToString(r.Hash.Sum(nil))
		if oid == r.ExpectedOid {
			return n, io.EOF
		}

		return n, fmt.Errorf("Expected OID: %s, got: %s", r.ExpectedOid, oid)
	}

	return n, nil
}
