package main

import "fmt"

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
