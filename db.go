package main

import "github.com/reusee/gobfile"

type File struct {
	Hash2m string
	Size   int64
}

type Db struct {
	file  *gobfile.File
	Files []File
}

func NewDb(path string) (db *Db, err error) {
	db = &Db{}
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
