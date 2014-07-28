package main

import (
	"fmt"
	"reflect"
	"sort"
)

var (
	p = fmt.Printf
	s = fmt.Sprintf
)

type T interface{}

func formatBytes(n int64) string {
	units := "bkmgtp"
	i := 0
	result := ""
	for n > 0 {
		res := n % 1024
		if res > 0 {
			result = fmt.Sprintf("%d%c", res, units[i]) + result
		}
		n /= 1024
		i++
	}
	if result == "" {
		return "0"
	}
	return result
}

func sortBy(s T, by T) {
	sort.Sort(&_Sorter{
		slice: reflect.ValueOf(s),
		cmp:   reflect.ValueOf(by),
	})
}

type _Sorter struct {
	slice reflect.Value
	cmp   reflect.Value
}

func (s *_Sorter) Len() int {
	return s.slice.Len()
}

func (s *_Sorter) Less(i, j int) bool {
	return s.cmp.Call([]reflect.Value{
		s.slice.Index(i),
		s.slice.Index(j),
	})[0].Bool()
}

func (s *_Sorter) Swap(i, j int) {
	tmp := reflect.ValueOf(s.slice.Index(i).Interface())
	s.slice.Index(i).Set(s.slice.Index(j))
	s.slice.Index(j).Set(tmp)
}
