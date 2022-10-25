package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"math/rand"
	"strings"
	"regexp"
	"sync"
)

// Map dimensions
const WIDTH = 15
const HEIGHT = 30

const LOGO = `
______   __    __  _______
/      \ /  |  /  |/       \
/$$$$$$  |$$ |  $$ |$$$$$$$  |
$$ | _$$/ $$ |  $$ |$$ |  $$ |
$$ |/    |$$ |  $$ |$$ |  $$ |
$$ |$$$$ |$$ |  $$ |$$ |  $$ |
$$ \__$$ |$$ \__$$ |$$ |__$$ |
$$    $$/ $$    $$/ $$    $$/
$$$$$$/   $$$$$$/  $$$$$$$/
`

/*
Singleton restricts instaniation of struct to a single instance
Singletons also provide a global access to an instance and protects it from being overwritten
*/
var once sync.Once
var worldInstance *worldSingle

func getWorldInstance() *worldSingle {
	if worldInstance == nil {
		once.Do(
			func() {
				worldInstance = &worldSingle{}
			})
	}

	return worldInstance
}

type worldSingle struct {
	worldMap [WIDTH][HEIGHT]int // Pixel array which stores the map to create a perfect square
	items []Item
}

// Items are located around the map (need to be generated on startup from data file)
type Item struct {
	description string
	coordinates Point
	isActive    bool
	itemType ItemType
}

// Type of an item is used to define the actions that can be taken with an item
type ItemType int
const (
	Armour ItemType = iota
	Weapon
	Random
	HotSpot
	NPC
)

// Dictionary called eventObject of strings to functions which transmute the player eventObjects
var eventObject = map[string]func(player Player, modifers []string) {
	"hotspot": func(player Player, modifers []string) {
		// Finding a hotspot moves the player to a random location within the dungeon
		fmt.Println("You have found a hotspot - prepare to be deported")

		// Destroy the hotspot as they are single use
		newPos := findFreeLocationInDungeon()
		player.coordinates = newPos
	},
	"npc": func(player Player, modifers []string) {
		// NPC's sell random items to user on their request
		writeToPlayer(player.conn, "Hi I am " + modifers[0] + "! What would you like to buy?")
		writeToPlayer(player.conn, "I sell the following items: ")

//		sellableItems := make([]Item, 0)

//		n := 0
//		for n < rand.Intn(len(items)) {
//			sellableItems = append(sellableItems, getWorldInstance().items[rand.Intn(len(items))])
//			n += 1
//		}
//
//		writeToPlayer(player.conn, Reduce(sellableItems, func (item Item, acc string) string {
//			return acc + item.description
//		}, ""))

		// TODO: Provide options to buy stuff
	},
}

func initaliseGame() {
	// Pick random start point within the array
	var point = Point{WIDTH / 2, HEIGHT / 2, 0, 0, nil}

	// Generate a number of walks to make an actual dungeon
	for i := 0; i < 50; i++ {
		// Pick random direction to walk
		direction := pickRandomDirection()

		// Walk that direction for a random amount
		cyclesToWalk := rand.Intn(15)

		for j := 0; j < cyclesToWalk; j++ {
			getNewPoint(direction, &point)

			// Tiles become 0 on default and 1 when walked on to generate dungeon rooms
			getWorldInstance().worldMap[point.x][point.y] = 1
		}
	}

	// Retrieve items which are predefined in a text file
	content, err := os.ReadFile("data/items.txt")
	if err != nil {
		panic(err)
	}

	itemNames := strings.Split(string(content), "\n")

	for _, itemName := range itemNames {
		// Find a random location within the dungeon to place the item which is free
		randomPoint := findFreeLocationInDungeon()
		itemType := Random

		// Check type
		if strings.Contains(itemName, "armour") {
			itemType = Armour
		} else if strings.Contains(itemName, "sword") || strings.Contains(itemName, "spear") {
			itemType = Weapon
		}

		// Push new item to global items array
		getWorldInstance().items = append(getWorldInstance().items, Item{itemName, *randomPoint, true, itemType })
	}

	// Generate a number of event objects located around the map
	for i := 0; i < rand.Intn(10); i++ {
		randomPoint := findFreeLocationInDungeon()
		getWorldInstance().items = append(getWorldInstance().items, Item{"hotspot", *randomPoint, true, HotSpot})
	}

	// Generate a number of NPC's which are able to sell items

	npcNames, err := os.ReadFile("data/npcNames.txt")
	if err != nil {
		panic(err)
	}

	for _, name := range strings.Split(string(npcNames), "\n") {
		getWorldInstance().items = append(getWorldInstance().items, Item{name, *findFreeLocationInDungeon(), true, NPC})
	}
}

func handleConnection(conn net.Conn) {
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	writeToPlayer(conn, LOGO)

	// Create new player and retrieve name
//	writeToPlayer(conn, "Good day fellow union member!")
//	writeToPlayer(conn, "By what do you wish to be addressed by?")
//
//	nameBytes := make([]byte, 256)
//	_, _ = conn.Read(nameBytes)

//	nameStr := nonAlphanumericRegex.ReplaceAllString(string(nameBytes), "")

	nameStr := "sid"

	inventory := make([]Item, 1)
	inventory[0] = Item{ "blonde", Point {15, 20, 0, 0, nil}, true, Random}

	player := NewPlayer(findFreeLocationInDungeon(), conn, nameStr)

	// Dictionary of actions which players can undertake
	var actions = map[string]func(modifiers []string){
		"move": player.move,
		"scan": player.scan,
		"investigate": player.investigate,
		"locate": player.locate,
		"pickup": player.pickup,
		"drop": player.drop,
		"combine": player.combine,
		"stats": player.viewStats,
		"equip": player.equip,
		"unequip": player.unequip,
		"quit": player.quit,
		"help": player.help,
	}

	player.actions = actions

	writeToPlayer(player.conn, "Welcome to GUD! " + nameStr)

	for true {
		// Parse commands a user enters
		tmp := make([]byte, 256)
		conn.Read(tmp)

		// Parse input by changing all special
		parsedInput := strings.Split(nonAlphanumericRegex.ReplaceAllString(string(bytes.Trim(tmp, "\x00")), ""), " ")

		if ContainsKey(actions, parsedInput[0]) {
			actions[parsedInput[0]](parsedInput[1:len(parsedInput)])
		} else {
			writeToPlayer(player.conn, "Unknown command - please refer to the help command")
		}
	}
}

func startServer() {
	port := "localhost:5000"
	ln, err := net.Listen("tcp", port) // Create a new server
	if err != nil {
		panic(err)
	}

	fmt.Println("GUD running on port " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(conn)
	}
}

func main() {
	initaliseGame()
	fmt.Println("Game loaded")
	startServer()
}


