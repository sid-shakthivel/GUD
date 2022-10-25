package main

import (
	"math/rand"
)

type Point struct {
	x int
	y int
	gcost int // Cost of start point to end goal
	hcost int // Heuristic cost estimated cost from node to goal
	parent *Point
}

func NewPoint(x int, y int) *Point {
	p := new(Point)
	p.x = x
	p.y = y
	return p
}

var directions = map[string]func(point *Point) {
	"north": func(point *Point) {
		point.y = min((*point).y+1, HEIGHT)
	},
	"south": func(point *Point) {
		point.y = max((*point).y-1, 0)
	},
	"east": func(point *Point) {
		point.x = min((*point).x+1, WIDTH)
	},
	"west": func(point *Point) {
		point.x = max((*point).x-1, 0)
	},
	"north-east": func(point *Point) {
		point.y = min((*point).y+1, HEIGHT)
		point.x = min((*point).x+1, WIDTH)
	},
	"north-west": func(point *Point) {
		point.y = min((*point).y+1, HEIGHT)
		point.x = max((*point).x-1, 0)
	},
	"south-east": func(point *Point) {
		point.y = max((*point).y-1, 0)
		point.x = min((*point).x+1, WIDTH)
	},
	"south-west": func(point *Point) {
		point.y = max((*point).y-1, 0)
		point.x = max((*point).x-1, 0)
	},
}

func pickRandomDirection() string {
	directionsList := GetKeys(directions)
	return directionsList[len(directionsList) - 1]
}
func findFreeLocationInDungeon() *Point {
	// Get random coordinate and ensure dungeon space exists there
	randX := rand.Intn(WIDTH)
	randY := rand.Intn(HEIGHT)

	pos := getWorldInstance().worldMap[randX][randY]

	if pos == 1 {
		return NewPoint(randX, randY)
	} else {
		return findFreeLocationInDungeon()
	}
}

func getNewPoint(direction string, point *Point) {
	directions[direction](point)
}
