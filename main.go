package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 2 MiB
const bufSize = 1 << 21

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: hashr <directory>")
	}

	var err error
	rawBuf := new([bufSize]byte)

	rootPath := os.Args[1]
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		start := time.Now()

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		// Init hashes
		dMd5 := md5.New()
		dSha1 := sha1.New()
		dSha256 := sha256.New()
		dSha512 := sha512.New()

		buf := rawBuf[:]
		for {
			// Read chunk
			var n int
			n, err = file.Read(buf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				return err
			}

			buf = buf[:n]

			// Hash rounds
			dMd5.Write(buf)
			dSha1.Write(buf)
			dSha256.Write(buf)
			dSha512.Write(buf)
		}

		// Finalize hashes
		sumMd5 := hex.EncodeToString(dMd5.Sum(nil))
		sumSha1 := hex.EncodeToString(dSha1.Sum(nil))
		sumSha256 := hex.EncodeToString(dSha256.Sum(nil))
		sumSha512 := hex.EncodeToString(dSha512.Sum(nil))

		// Escape quotes in filename
		key := strings.Replace(path, `"`, `\"`, -1)

		log.Printf("Done %s in %s.", path, time.Since(start))
		fmt.Printf(`SET "%s" "%s|%s|%s|%s"`+"\n",
			key,
			sumMd5,
			sumSha1,
			sumSha256,
			sumSha512,
		)

		// No error
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
