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
	cmp := func(l, r *PathInfo) bool { // sort by path as default
		return l.path < r.path
	}
	var filters []func(*PathInfo) bool
	for _, arg := range args {
		switch arg {
		case "d", "date": // sort by date
			cmp = func(l, r *PathInfo) bool {
				return l.ModTime.After(r.ModTime)
			}
		case "s", "size": // sort by size
			cmp = func(l, r *PathInfo) bool {
				return l.file.Size > r.file.Size
			}
		case "w", "watch": // sort by watch count
			cmp = func(l, r *PathInfo) bool {
				return l.file.WatchCount > r.file.WatchCount
			}
		case "t", "tag": // sort by tag count
			cmp = func(l, r *PathInfo) bool {
				return len(l.file.Tags) > len(r.file.Tags)
			}
		case "rev", "reverse": // reverse sort
			c := cmp
			cmp = func(l, r *PathInfo) bool {
				return !c(l, r)
			}
		case "r", "random": // random sort
			cmp = func(l, r *PathInfo) bool {
				return rand.Intn(2) == 0
			}
		default:
			kw := arg
			filters = append(filters, func(info *PathInfo) bool {
				for _, tag := range info.file.Tags {
					parts := strings.Split(tag.Description, " ")
					for _, t := range parts {
						if kw == strings.TrimSpace(t) {
							return true
						}
					}
				}
				return strings.Contains(strings.ToLower(info.path), kw)
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
