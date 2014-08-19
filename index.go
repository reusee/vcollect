package main

import (
	"crypto/sha512"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var videoExtensions = []string{
	".rmvb",
}

func (db *Db) index() {
	t0 := time.Now()

	hashes := make(map[string]string)
	hasher := sha512.New()

	hashed := make(map[string]int)
	for index, f := range db.Files {
		hashed[s("%s-%d", f.Hash2m, f.Size)] = index
	}

	filepath.Walk(filepath.Dir(db.path), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		mimeType := mime.TypeByExtension(ext)
		if !strings.HasPrefix(mimeType, "video") {
			for _, e := range videoExtensions {
				if e == ext {
					goto is_video
				}
			}
			return nil
		}
	is_video:
		p("%s\n", path)

		hasher.Reset()
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		io.CopyN(hasher, f, 1024*1024*2)
		//h := string(hasher.Sum(nil))
		h := s("%x", hasher.Sum(nil))
		if name, has := hashes[h]; has { // duplicated file or conflict
			p("=== duplicated file or hash conflict. stop processing ===\n")
			p("%s\n", name)
			p("%s\n", path)
			panic("stop")
		}
		hashes[h] = path

		// update file info
		index, has := hashed[s("%s-%d", h, info.Size())]
		if !has { // new file
			db.Files = append(db.Files, &FileInfo{
				Hash2m: h,
				Size:   info.Size(),
			})
			index = len(db.Files) - 1
		}
		// add file paths
		fileinfo := db.Files[index]
		has = false
		for _, p := range fileinfo.Filepaths {
			if p == path {
				has = true
				break
			}
		}
		if !has {
			fileinfo.Filepaths = append(fileinfo.Filepaths, path)
		}

		// update path info
		pathInfo, ok := db.Paths[path]
		if !ok {
			pathInfo = new(PathInfo)
			db.Paths[path] = pathInfo
		}
		pathInfo.Index = index
		pathInfo.ModTime = info.ModTime()

		return nil
	})
	p("=== %d files indexed ===\n", len(db.Files))
	p("%v\n", time.Now().Sub(t0))

	// clear paths
	p("=== clear path ===\n")
	for path, _ := range db.Paths {
		_, err := os.Stat(path)
		if err != nil {
			delete(db.Paths, path)
			p("%s\n", path)
		}
	}

	err := db.Save()
	if err != nil {
		panic(err)
	}

}
