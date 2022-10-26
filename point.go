package main

import (
	"math/rand"
	"strconv"
)

type Point struct {
	x int
	y int
	gcost int // Cost of start point to end goal
	hcost int // Heuristic cost estimated cost from node to goal
	parent *Point
}

func (point Point) format() string {
	return "[" + strconv.Itoa(point.x) + "," + strconv.Itoa(point.y) + "]"
}

func NewPoint(x int, y int) *Point {
	p := new(Point)
	p.x = x
	p.y = y
	return p
}

var manhattanDirections = map[string]func(point *Point) {
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
	"northeast": func(point *Point) {
		point.y = min((*point).y+1, HEIGHT)
		point.x = min((*point).x+1, WIDTH)
	},
	"northwest": func(point *Point) {
		point.y = min((*point).y+1, HEIGHT)
		point.x = max((*point).x-1, 0)
	},
	"southeast": func(point *Point) {
		point.y = max((*point).y-1, 0)
		point.x = min((*point).x+1, WIDTH)
	},
	"southwest": func(point *Point) {
		point.y = max((*point).y-1, 0)
		point.x = max((*point).x-1, 0)
	},
}

func pickPerpendicularRandomDirection(lastDirection string) string {
	directionsList := GetKeys(manhattanDirections)
	newDirection := directionsList[len(directionsList) - 1]

	if lastDirection == "north" && newDirection == "south" || lastDirection == "north" && newDirection == "north" {
		return pickPerpendicularRandomDirection(lastDirection)
	}

	if lastDirection == "south" && newDirection == "north" || lastDirection == "south" && newDirection == "south" {
		return pickPerpendicularRandomDirection(lastDirection)
	}

	if lastDirection == "east" && newDirection == "east" || lastDirection == "east" && newDirection == "west" {
		return pickPerpendicularRandomDirection(lastDirection)
	}

	if lastDirection == "west" && newDirection == "east" || lastDirection == "west" && newDirection == "west" {
		return pickPerpendicularRandomDirection(lastDirection)
	}

	return newDirection
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
