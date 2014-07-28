package main

import (
	"time"

	"github.com/reusee/jsonfile"
)

type FileInfo struct {
	Hash2m string
	Size   int64
	Tags   []*Tag
}

type Tag struct {
	Description string
	Position    int64
}

func (i *FileInfo) AddTag(pos int64, desc string) {
	i.Tags = append(i.Tags, &Tag{
		Description: desc,
		Position:    pos,
	})
}

type PathInfo struct {
	Index   int
	ModTime time.Time

	path string
	file *FileInfo
}

type Db struct {
	file  *jsonfile.File
	path  string
	Files []*FileInfo
	Paths map[string]*PathInfo
}

func NewDb(path string) (db *Db, err error) {
	db = &Db{
		path:  path,
		Paths: make(map[string]*PathInfo),
	}
	db.file, err = jsonfile.New(db, path, 52218)
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
