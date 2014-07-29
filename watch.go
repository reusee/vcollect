package main

import (
	"math/rand"
	"strings"
)

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
	var filters []func(*PathInfo) bool
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
		default:
			kw := arg
			filters = append(filters, func(info *PathInfo) bool {
				if strings.Contains(info.path, kw) {
					return true
				}
				return false
			})
		}
	}

	// sort
	sortBy(infos, cmp)

	// filter
	if len(filters) > 0 {
		unfiltered := infos
		infos = make([]*PathInfo, 0)
		for _, info := range unfiltered {
			ok := false
			for _, f := range filters {
				if f(info) {
					ok = true
				}
			}
			if ok {
				infos = append(infos, info)
			}
		}
	}

	if len(infos) == 0 {
		p("no media found.\n")
		return
	}
	db.lgi_gst(infos)
}
