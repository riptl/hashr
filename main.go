package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 2 MiB
const bufSize = 1 << 21

var rootPath string
var workers int
var syncer = int64(1 << 62)
var printLock sync.Mutex
var prefix string

func main() {
	flag.IntVar(&workers, "threads", runtime.NumCPU(), "Number of files to process at simultaneously")
	flag.StringVar(&prefix, "prefix", "", "Path prefix")

	if len(os.Args) < 2 {
		log.Fatal("Usage: hashr <directory>")
	} else if strings.Contains(os.Args[1], "help") {
		os.Args[0] = os.Args[0] + " <directory>"
		flag.Usage()
		os.Exit(1)
	}

	parseFlags(os.Args[2:])

	var err error
	jobs := make(chan string)

	for i := 0; i < workers; i++ {
		go worker(jobs)
	}

	rootPath = os.Args[1]
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Error accessing %s\n", path)
		} else if !info.IsDir() {
			atomic.AddInt64(&syncer, 1)
			jobs <- path
		}
		return nil
	})

	// Mark files walking as done
	atomic.AddInt64(&syncer, -(1 << 62))

	for atomic.LoadInt64(&syncer) != 0 {
		time.Sleep(time.Millisecond)
	}

	close(jobs)

	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags(cmd []string) {
	realArgs := os.Args
	os.Args = append([]string{"hashr"}, cmd...)
	flag.Parse()
	os.Args = realArgs
}

func worker(jobs <-chan string) {
	buf := make([]byte, bufSize)
	for job := range jobs {
		hashes(job, buf)
		atomic.AddInt64(&syncer, -1)
	}
}

func hashes(path string, buf []byte) {
	fullBuf := buf
	start := time.Now()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		log.Printf("Error processing %s: %e\n", path, err)
		return
	}

	// Init hashes
	dMd5 := md5.New()
	dSha1 := sha1.New()
	dSha256 := sha256.New()
	dSha512 := sha512.New()

	eof := false
	for !eof {
		// Read chunk
		var n int
		n, err = file.Read(buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			eof = true
		} else if err != nil {
			log.Printf("Error processing %s: %e\n", path, err)
			return
		}

		buf = buf[:n]

		// Hash rounds
		dMd5.Write(buf)
		dSha1.Write(buf)
		dSha256.Write(buf)
		dSha512.Write(buf)

		buf = fullBuf
	}

	// Finalize hashes
	sumMd5 := hex.EncodeToString(dMd5.Sum(nil))
	sumSha1 := hex.EncodeToString(dSha1.Sum(nil))
	sumSha256 := hex.EncodeToString(dSha256.Sum(nil))
	sumSha512 := hex.EncodeToString(dSha512.Sum(nil))

	// Escape quotes in filename
	keyPath, _ := filepath.Rel(rootPath, path)
	keyPath = filepath.Join(prefix, keyPath)
	key := strings.Replace(keyPath, `"`, `\"`, -1)

	log.Printf("Done %s in %s.", keyPath, time.Since(start))

	// Print result
	printLock.Lock()
	fmt.Printf(`SET "%s" "%s|%s|%s|%s"`+"\n",
		key,
		sumMd5,
		sumSha1,
		sumSha256,
		sumSha512,
	)
	printLock.Unlock()
}
