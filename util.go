package main

import (
	"net"
	"unicode"
	"math"
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

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
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
	conn.Write([]byte(text + "\n\n"))
}

func calculateHeuristicCost(nodeA Point, nodeB Point) int {
	/*
	Uses Manhattan distance in which we check nodes horizontally and vertically (not diagonally) - named because it's similar to calculating number of city blocks
	Delta X + Delta Y
	*/

	deltaX := int(math.Abs(float64(nodeA.x - nodeB.x)))
	deltaY := int(math.Abs(float64(nodeA.y - nodeB.y)))

	if deltaX > deltaY {
		return 14 * deltaY + 10 * (deltaX - deltaY)
	} else {
		return 14 * deltaX + 10 * (deltaY - deltaX)
	}
}