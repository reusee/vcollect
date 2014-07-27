package main

import (
	"crypto/sha512"
	"encoding/gob"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	p = fmt.Printf
	s = fmt.Sprintf
)

func init() {
	gob.Register(new(Db))
}

func main() {
	// path
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if !stat.IsDir() {
		panic("not dir")
	}

	// db
	db, err := NewDb(filepath.Join(path, ".vcollect"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// indexing
	hashes := make(map[string]string)
	hasher := sha512.New()
	tt0 := time.Now()
	hashed := make(map[string]bool)
	for _, f := range db.Files {
		hashed[s("%s-%d", f.Hash2m, f.Size)] = true
	}
	var files []File
	filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
		if !strings.HasPrefix(mimeType, "video") {
			return nil
		}
		p("%s\n", path)
		t0 := time.Now()
		hasher.Reset()
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		io.CopyN(hasher, f, 1024*1024*2)
		h := string(hasher.Sum(nil))
		if name, has := hashes[h]; has { // duplicated file or conflict
			p("=== duplicated file or hash conflict. stop processing ===\n")
			p("%s\n", name)
			p("%s\n", path)
			panic("stop")
		}
		hashes[h] = path
		p("%s %v\n", formatBytes(info.Size()), time.Now().Sub(t0))
		if !hashed[s("%s-%d", h, info.Size())] {
			files = append(files, File{
				Hash2m: h,
				Size:   info.Size(),
			})
		}
		return nil
	})
	// add to db
	db.Files = append(db.Files, files...)
	p("=== %d files indexed ===\n", len(db.Files))
	p("%v\n", time.Now().Sub(tt0))

	err = db.Save()
	if err != nil {
		panic(err)
	}
}
