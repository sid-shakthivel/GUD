package main

import (
	"fmt"
	"math/rand"
)

type Point struct {
	x int
	y int
	gcost int // Cost of start point to end goal
	hcost int // Heuristic cost estimated cost from node to goal
	parent *Point
}

type Directions int

const (
	North Directions = iota
	South
	East
	West
	NorthWest
	NorthEast
	SouthWest
	SouthEast
)

func pickRandomDirection() Directions {
	switch rand.Intn(8) + 0 {
	case 0:
		return North
	case 1:
		return South
	case 2:
		return East
	case 3:
		return West
	case 4:
		return NorthWest
	case 5:
		return NorthEast
	case 6:
		return SouthWest
	case 7:
		return SouthEast
	default:
		return West
	}
}

func printMap() {
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			if getWorldInstance().worldMap[i][j] == 1 {
				fmt.Print("#")
			} else {
				fmt.Print("~")
			}
		}
		fmt.Println()
	}
}

func findFreeLocationInDungeon() Point {
	// Get random coordinate and ensure dungeon space exists there
	randX := rand.Intn(WIDTH)
	randY := rand.Intn(HEIGHT)

	pos := getWorldInstance().worldMap[randX][randY]

	if pos == 1 {
		return Point{randX, randY, 0, 0, nil}
	} else {
		return findFreeLocationInDungeon()
	}
}

func calculateNewPoint(direction Directions, point *Point) {
	switch direction {
	case North:
		(*point).y = min((*point).y+1, HEIGHT)
	case South:
		(*point).y = max((*point).y-1, 0)
	case East:
		(*point).x = min((*point).x+1, WIDTH)
	case West:
		(*point).x = max((*point).x-1, 0)
	case NorthWest:
		calculateNewPoint(North, point)
		calculateNewPoint(West, point)
	case NorthEast:
		calculateNewPoint(North, point)
		calculateNewPoint(East, point)
	case SouthWest:
		calculateNewPoint(South, point)
		calculateNewPoint(West, point)
	case SouthEast:
		calculateNewPoint(South, point)
		calculateNewPoint(East, point)
	}
}
