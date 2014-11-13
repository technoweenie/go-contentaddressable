/*
Package contentaddressable contains tools for reading and writing content
addressable files. Files are written to a temporary location, and only renamed
to the final location after the file's OID (Object ID) has been verified.

    filename := "path/to/01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b"
    file, err := contentaddressable.NewFile(filename)
    if err != nil {
      panic(err)
    }
    defer file.Close()

    file.Oid // 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b

    written, err := io.Copy(file, someReader)

    if err == nil {
    // Move file to final location if OID is verified.
      err = file.Accept()
    }

    if err != nil {
      panic(err)
    }

Currently SHA-256 is used for a file's OID.

You can also read files, while verifying that they are not corrupt.

    filename := "path/to/01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b"

    // get this from doing an os.Stat() or something
    expectedSize := 123

    // returns a contentaddressable.ReadCloser, with some extra functions on top
    // of io.ReadCloser.
    reader, err := contentaddressable.Open(filename)
    if err != nil {
      panic(err)
    }
    defer file.Close()

    written, err := io.Copy(ioutil.Discard, reader)
    if err != nil {
      panic(err)
    }

    seenBytes := reader.SeenBytes()

    if written != seenBytes {
      panic("reader is broken")
    }

    if seenBytes < expected {
      panic("partial read")
    }

    if reader.Oid() != "01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b" {
      panic("SHA-256 signature doesn't match expected")
    }
*/
package contentaddressable
