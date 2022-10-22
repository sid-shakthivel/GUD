package main

import (
	"net"
	"fmt"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return 0
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return 0
}

func Find[T any] (s []T, f func(T) bool) int {
	for i := range s {
		if f(s[i]) == true {
			return i
		}
	}
	return -1
}

/*
	Signature:
	Reduce (slice, func(current, acc), inital)
*/
func Reduce[T, U any](s []T, f func(T, U) U, initValue U) U {
	acc := initValue
	for _, v := range s {
		acc = f(v, acc)
	}
	return acc
}

func Contains[T comparable] (s []T, e T) bool {
	for i := range s {
		if s[i] == e {
			return true
		}
	}
	return false
}

func RemoveAtIndex[T any] (s []T, index int) []T {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

func GetKeys[T comparable, U any] (m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k, _ := range m {
		fmt.Println(k)
		keys = append(keys, k)
	}
	return keys
}

func ContainsKey[T comparable, U any] (m map[T]U, s T) bool {
	for k, _ := range m {
		if k == s {
			return true
		}
	}
	return false
}

func GetValues[T comparable, U any] (m map[T]U) [] U {
	values := make([]U, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func writeToPlayer(conn net.Conn, text string) {
	conn.Write([]byte("$ " + text + "\n\n"))
}
