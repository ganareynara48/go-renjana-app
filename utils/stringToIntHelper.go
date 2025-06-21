package utils

import "strconv"

func SafeAtoi(s string) (int, error) {
	return strconv.Atoi(s)
}
