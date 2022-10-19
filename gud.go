package main

/*
	Singleton restricts instaniation of struct to a single instance
	Singletons also provide a global access to an instance and protects it from being overwritten
*/

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sort"
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
		case "north":
			if player.isWithinBoundry(Point{player.coordinates.x, player.coordinates.y + 1, 0, 0, nil }) {
				player.coordinates.y = player.coordinates.y + 1
			}
		case "south":
			if player.isWithinBoundry(Point{player.coordinates.x, player.coordinates.y - 1, 0, 0, nil}) {
				player.coordinates.y = player.coordinates.y - 1
			}
		case "east":
			if player.isWithinBoundry(Point{player.coordinates.x + 1, player.coordinates.y, 0, 0, nil}) {
				player.coordinates.x = player.coordinates.x + 1
			}
		case "west":
			if player.isWithinBoundry(Point{player.coordinates.x - 1, player.coordinates.y, 0, 0, nil}) {
				player.coordinates.x = player.coordinates.x - 1
			}
		default:
			panic("Random direction")
		}
	}
}

func (player Player) isWithinBoundry(newCoordinate Point) bool {
	if newCoordinate.x > 20 || newCoordinate.y >= 40 || world[newCoordinate.x][newCoordinate.y] == 0 {
		return false
	}
	return true
}

/*
	Signature `scan {distance}`
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
	Signature `investigate {item}`
	Allows player to investigate items they encounter including hotspots
*/
func (player Player) investigate(modifiers []string) {
	// Check if item is within the eventObject dictionary and run functions
	eventObject[modifiers[0]](player)
}

// Check if a point array contains a point with the same coordiantes
func (point Point) ContainsPoint(points []Point) bool {
	for i := range points {
		if points[i].x == point.x && points[i].y == point.y {
			return true
		}
	}
	return false
}

/*
	Signature 'locate {item name}`
	Uses Dijkstra's path finding algorithm to work out the shortest path between the user and an item
	Dispays the path to the user
*/
func (player Player) locate(modifiers []string) []Point {
	// Find item position in world
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex > -1 {
		itemPosition := player.inventory[itemIndex].coordinates

		fmt.Println("Item Position is ", itemPosition)
		fmt.Println("Player Position is ", player.coordinates)

		var openNodes []Point // Nodes that have calculated cost
		var closedNodes []Point // Nodes that haven't calculated cost

		openNodes = append(openNodes, player.coordinates) // Add starting node

		for len(openNodes) > 0 {
			// Sort the open nodes to get the one with the lowest heuristic cost (cost to the actual node)
			sort.SliceStable(openNodes, func(i, j int) bool {
				return openNodes[i].hcost < openNodes[j].hcost
			})

			// Acknowledge the current node has been accounted for and use it
			currentNode := openNodes[0]
			closedNodes = append(closedNodes, currentNode)
			openNodes = RemoveAtIndex(openNodes, 0)

			// If we have found the target, alert the user (for now)

			if currentNode.x == itemPosition.x && currentNode.y == itemPosition.y {
				path := make([]Point, 0)

				node := currentNode
				path = append(path, node)

				for node.parent != nil {
					node = *node.parent
					path = append(path, node)
				}

				return path
			} else {
				// Create a list of adjacent nodes which are walkable from the current node and not closed

				neighbour1 := Point{min(currentNode.x + 1, WIDTH - 1), currentNode.y, 0, 0, nil }
				neighbour2 := Point{max(currentNode.x - 1, 0), currentNode.y, 0, 0, nil}
				neighbour3 := Point{currentNode.x, min(currentNode.y + 1, HEIGHT - 1), 0, 0, nil}
				neighbour4 := Point{currentNode.x, max(currentNode.y - 1, 0), 0, 0, nil}
				neighbour5 := Point{min(currentNode.x + 1, WIDTH - 1), min(currentNode.y + 1, HEIGHT - 1), 0, 0, nil }
				neighbour6 := Point{min(currentNode.x + 1, WIDTH -1), max(currentNode.y - 1, 0), 0, 0, nil}
				neighbour7 := Point{max(currentNode.x - 1, 0), min(currentNode.y + 1, HEIGHT - 1), 0, 0, nil}
				neighbour8 := Point{max(currentNode.x - 1, 0), max(currentNode.y - 1, 0), 0, 0, nil}

				neighours := [8]Point { neighbour1, neighbour2, neighbour3, neighbour4, neighbour5, neighbour6, neighbour7, neighbour8 }

				for _, neighbour := range neighours {
					// Check if it's walkable (world[neighbour.x][neighbour.y] == 1) and not on the closed list
					if !neighbour.ContainsPoint(closedNodes) {
						cost := calculateHeuristicCost(currentNode, neighbour) + currentNode.gcost

						if cost < neighbour.gcost || !neighbour.ContainsPoint(openNodes) {
							neighbour.gcost = cost
							neighbour.hcost = calculateHeuristicCost(neighbour, itemPosition)
							neighbour.parent = &currentNode
						}

						if !neighbour.ContainsPoint(openNodes) {
							openNodes = append(openNodes, neighbour)
						}
					}
				}
			}
		}
	}
	println("Lol failed")

	return make([]Point, 0)
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

		// Possibly should check if items can be combined - make a random method for that maybe
	} else {
		panic("User does not possess both items")
	}
}

func Find[T any] (s []T, f func(T) bool) int {
	for i := range s {
		if f(s[i]) == true {
			return i
		}
	}
	return -1
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

func GetValues[T comparable, U any] (m map[T]U) [] U {
	values := make([]U, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Map dimensions
const WIDTH = 20
const HEIGHT = 40

// Create a 20 * 40 pixel array which stores the map to create a perfect square
var world [WIDTH][HEIGHT]int

// Global list of items which are present within the cave
var items []Item

// Global list of strings which map to eventObject dictionary
var eventObjects []string

func move(modifiers []string) {
	fmt.Println("We move'd")
}

// Create dictionary called actions which players can undertake
var actions = map[string]func(modifiers []string){
	"move": move,
}

/*
	Dictionary called eventObject of strings to functions which transmute the player eventObjects
*/
var eventObject = map[string]func(state Player) {
	"hotspot": func(state Player) {
		// Finding a hotspot will trigger this method which moves the player to a random location

		fmt.Println("You have found a hotspot - prepare to be deported")
		newPos := findFreeLocationInDungeon()
		state.coordinates = newPos
	},
}
type Point struct {
	x int
	y int
	gcost int // Cost of start point to end goal
	hcost int // Heuristic estimated cost from node to goal
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
		return Point{randX, randY, 0, 0, nil}
	} else {
		return findFreeLocationInDungeon()
	}
}

func initaliseGame() {
	// Pick random start point within the array
	var point = Point{WIDTH / 2, HEIGHT / 2, 0, 0, nil}

	// Generate a number of walks to make an actual dungeon
	for i := 0; i < 50; i++ {
		// Pick random direction to walk
		direction := pickRandomDirection()

		// Walk that direction for a random amount
		cyclesToWalk := rand.Intn(20)

		for j := 0; j < cyclesToWalk; j++ {
			calculateNewPoint(direction, &point)

			// Tiles become 0 on default and 1 when walked on to generate dungeon rooms
			world[point.x][point.y] = 1
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

	// Generate a number of event objects located around the map
	for i := 0; i < rand.Intn(10); i++ {
		randomPoint := findFreeLocationInDungeon()
		items = append(items, Item{GetKeys(eventObject)[rand.Intn(len(eventObject))], randomPoint, true})
	}

	printMap()
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
//	actions["move"]([]string{})

	args := make([]string, 1)
	inventory := make([]Item, 1)
	args[0] = "blonde"
	inventory[0] = Item{ "blonde", Point {15, 20, 0, 0, nil}, true}

	player := Player{findFreeLocationInDungeon(), inventory}

	fmt.Println(player.coordinates)

	player.locate(args)

	fmt.Println("Finished execution")

	startServer()
}

