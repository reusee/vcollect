package main

import (
	"encoding/gob"
	"time"

	"github.com/reusee/gobfile"
)

func init() {
	gob.Register(new(Db))
}

type FileInfo struct {
	Hash2m string
	Size   int64
}

type PathInfo struct {
	Index   int
	ModTime time.Time

	path string
	file *FileInfo
}

type Db struct {
	file  *gobfile.File
	path  string
	Files []*FileInfo
	Paths map[string]*PathInfo
}

func NewDb(path string) (db *Db, err error) {
	db = &Db{
		path:  path,
		Paths: make(map[string]*PathInfo),
	}
	db.file, err = gobfile.New(db, path, 52218)
	if err != nil {
		return nil, err
	}
	return
}

func (d *Db) Save() error {
	return d.file.Save()
}

func (d *Db) Close() {
	d.file.Close()
}
