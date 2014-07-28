package main

import (
	"os"
	"path/filepath"
)

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
	db, err := NewDb(filepath.Join(path, ".vcollect.json"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// command
	if len(os.Args) <= 2 {
		p("usage: %s [path] [command] [args...]\n", os.Args[0])
		return
	}
	switch os.Args[2] {
	case "i", "index":
		db.index()
	case "w", "watch":
		db.watch(os.Args[3:])

	case "clean":
		for _, file := range db.Files {
			var tags []*Tag
			for _, tag := range file.Tags {
				if tag.Description != "" {
					tags = append(tags, tag)
				}
			}
			file.Tags = tags
		}
		db.Save()
	}
}
