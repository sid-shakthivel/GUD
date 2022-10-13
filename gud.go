package main

/*
	Singleton restricts instaniation of struct to a single instance
	Singletons also provide a global access to an instance and protects it from being overwritten
*/

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

// Items are located around the map (need to be generated on startup)
type Item struct {
	description string
	coordinates Point
	isActive    bool
}

// Players serve as clients too and connections to the server!?
type Player struct {
	coordinates Point
	inventory   []Item
}

/*
	Signature `move {direction} {distance}`
	Player coordinates are manipulated in a direction until they hit distance or a wall
*/
func (player Player) move(modifiers []string) {
	// Move in direction until we hit a wall or not

	distance, err := strconv.Atoi(modifiers[1])
	if err != nil {
		panic(err)
	}

	for i := 0; i < distance; i++ {
		switch modifiers[0] {
		case "North":
			if player.isWithinBoundry(Point{player.coordinates.x, player.coordinates.y.(int) + 1}) {
				player.coordinates.y = player.coordinates.y.(int) + 1
			}
		case "South":
			if player.isWithinBoundry(Point{player.coordinates.x, player.coordinates.y.(int) - 1}) {
				player.coordinates.y = player.coordinates.y.(int) - 1
			}
		case "East":
			if player.isWithinBoundry(Point{player.coordinates.x.(int) + 1, player.coordinates.y}) {
				player.coordinates.x = player.coordinates.x.(int) + 1
			}
		case "West":
			if player.isWithinBoundry(Point{player.coordinates.x.(int) - 1, player.coordinates.y}) {
				player.coordinates.x = player.coordinates.x.(int) - 1
			}
		default:
			panic("Random direction")
		}
	}
}

func (player Player) isWithinBoundry(newCoordinate Point) bool {
	if newCoordinate.x.(int) > 20 || newCoordinate.y.(int) >= 40 || world[newCoordinate.x.(int)][newCoordinate.y.(int)] == 0 {
		return false
	}
	return true
}

/*
	Signature `scan {distnace}`
	Allows player to scan nearby to idenify items and teleportation in all directions for a unit
*/
func (player Player) scan(modifiers []string) {
	// Looks for unit block in all directions to check for item and reports back to user

	distance, err := strconv.Atoi(modifiers[1])
	if err != nil {
		panic(err)
	}

	for i := 0; i < distance; i++ {
		// Check for each direction
	}
}

/*
	Signature `pickup {item name}`
	Add item to inventory
	TODO: Possible to make some sort of method to combine these two
*/
func (player Player) pickup(modifiers []string) {
	// Search items array for item requested to get item of description at coordinate

	 itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	 })

	 if itemIndex < 0 {
	 	fmt.Println("Item not found")
	 } else {
		 player.inventory = append(player.inventory, items[itemIndex])
		 items = RemoveAtIndex(items, itemIndex)
	 }
}

/*
	Signature `drop {item name}`
	Remove item from inventory and place at a coordinate
*/
func (player Player) drop(modifiers []string) {
	// Search inventory for the item, remove it from inventory and append to items global
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		fmt.Println("Item not found")
	} else {
		items = append(items, player.inventory[itemIndex])
		player.inventory = RemoveAtIndex(player.inventory, itemIndex)
	}
}

/*
	Signature `combine {item1 name} {item2 name}
	Combines item to solve puzzles (when included)
*/
func (player Player) combine(modifiers []string) {
	// Check user hasn't entered same item twice
	if modifiers[0] == modifiers[1] {
		panic("Same name chosen - use connection to write message")
	}

	firstItemPosition := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	secondItemPosition := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[1]
	})

	// Check user has both items within their inventory
	if firstItemPosition > 0 && secondItemPosition > 0 {
		// Add a new combined item to inventory
		player.inventory = append(player.inventory, Item {player.inventory[firstItemPosition].description + player.inventory[secondItemPosition].description, player.inventory[firstItemPosition].coordinates, true })

		// Remove both items from players inventory
		RemoveAtIndex(player.inventory, firstItemPosition)
		RemoveAtIndex(player.inventory, secondItemPosition)

		// Possibly should check if items can be combined - make a random method for that
	} else {
		panic("User does not possess both items")
	}
}

// Generic function which returns the index of a slice entry
func Find[T any](s []T, f func(T) bool) int {
	for i := range s {
		if f(s[i]) == true {
			return i
		}
	}
	return -1
}

func RemoveAtIndex[T any] (s []T, index int) []T {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

// Map dimensions
const WIDTH = 20
const HEIGHT = 40

// Create a 20 * 40 pixel array which stores the map to create a perfect square
var world [WIDTH][HEIGHT]int

// Create a global list of items which are present within the cave (may or may not modify after init)
var items []Item

func move(modifiers []string) {
	fmt.Println("We move'd")
}

// Create dictionary actions which players can undertake
var actions = map[string]func(modifiers []string){
	"move": move,
}

type Point struct {
	x, y interface{}
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
	for i := 0; i < 20; i++ {
		for j := 0; j < 40; j++ {
			if world[i][j] == 1 {
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

	pos := world[randX][randY]

	if pos == 1 {
		return Point{randX, randY}
	} else {
		return findFreeLocationInDungeon()
	}
}

func initaliseGame() {
	// Pick random start point within the array
	var point = Point{WIDTH / 2, HEIGHT / 2}

	// Generate a number of walks to make an actual dungeon
	for i := 0; i < 50; i++ {
		// Pick random direction to walk
		direction := pickRandomDirection()

		// Walk that direction for a random amount
		cyclesToWalk := rand.Intn(20)

		for j := 0; j < cyclesToWalk; j++ {
			calculateNewPoint(direction, &point)

			// Tiles become 0 on default and 1 when walked on to generate dungeon rooms
			world[point.x.(int)][point.y.(int)] = 1
		}
	}

	// Identify items which are predefined in a text file
	content, err := os.ReadFile("items.txt")
	if err != nil {
		panic(err)
	}
	for i, item := range strings.Split(string(content), " ") {
		// Find a random location within the dungeon to place the item which is free
		randomPoint := findFreeLocationInDungeon()
		fmt.Println(i, item)

		// Push new item to global items array
		items = append(items, Item{item, randomPoint, true})
	}

	printMap()
}

func calculateNewPoint(direction Directions, point *Point) {
	switch direction {
	case North:
		(*point).y = min((*point).y.(int)+1, HEIGHT)
	case South:
		(*point).y = max((*point).y.(int)-1, 0)
	case East:
		(*point).x = min((*point).x.(int)+1, WIDTH)
	case West:
		(*point).x = max((*point).x.(int)-1, 0)
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

func handleConnection(conn net.Conn) {
	fmt.Println("Got a connection?")

	for true {
		// Read user data
		tmp := make([]byte, 256)
		conn.Read(tmp)

		// Parse commands
		switch string(bytes.Trim(tmp, "\x00")) {
		case "hello\n":
			conn.Write([]byte("Hello there new user\n"))
		case "clean\n":
			conn.Write([]byte("Cleaning room\n"))
		case "map\n":
			for i := 0; i < WIDTH; i++ {
				for j := 0; j < HEIGHT; j++ {
					if world[i][j] == 1 {
						conn.Write([]byte("#"))
					} else {
						conn.Write([]byte("~"))
					}
				}
				conn.Write([]byte("\n"))
			}
		default:
			conn.Write([]byte("Unknown command\n"))
		}
	}
}

func startServer() {
	port := "localhost:5000"
	ln, err := net.Listen("tcp", port) // Create a new server
	if err != nil {
		panic(err)
	}

	fmt.Println("Server is running on port " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(conn)
	}
}

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

func main() {
	initaliseGame()
	actions["move"]([]string{})
	startServer()
}

