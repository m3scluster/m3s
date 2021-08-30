package util

import "strings"

// SplitString will split a string or give just one value back
// src 	  = the string to split
// return = array of all strings
func SplitString(src string) []string {
	var des []string
	if strings.Contains(src, ",") {
		des = strings.Split(src, ",")
	} else {
		des[0] = src
	}
	return des
}
