package main

import "math/rand"

func (db *Db) watch(args []string) {
	var infos []*PathInfo
	for path, info := range db.Paths {
		info.path = path
		info.file = db.Files[info.Index]
		infos = append(infos, info)
	}

	// parse args
	cmp := func(l, r *PathInfo) bool {
		return l.path < r.path
	}
	for _, arg := range args {
		switch arg {
		case "n", "new":
			cmp = func(l, r *PathInfo) bool {
				return l.ModTime.After(r.ModTime)
			}
		case "o", "old":
			cmp = func(l, r *PathInfo) bool {
				return l.ModTime.Before(r.ModTime)
			}
		case "b", "big":
			cmp = func(l, r *PathInfo) bool {
				return l.file.Size > r.file.Size
			}
		case "s", "small":
			cmp = func(l, r *PathInfo) bool {
				return l.file.Size < r.file.Size
			}
		case "r", "random":
			cmp = func(l, r *PathInfo) bool {
				if rand.Intn(2) == 0 {
					return true
				}
				return false
			}
		}
	}

	// sort
	sortBy(infos, cmp)

	//db.qt(infos)
	//db.lgi_mplayer(infos)
	db.lgi_gst(infos)
}
