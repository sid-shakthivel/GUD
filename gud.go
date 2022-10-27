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
const WIDTH = 30 // Width of map
const HEIGHT = 15 // Height of map
const MAX_TUNNELS = 50 // Greatest number of turns algorithm can make
const MAX_TUNNEL_LENGTH = 30 // Greatest length of each tunnel the algorithm will choose before making a turn

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
	EventObject
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

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
		writeToPlayer(player.conn, "Goodday fellow union member I am " + modifers[0] + "! What would you like to buy?")
		writeToPlayer(player.conn, "I sell the following items: ")

		allItemsPresent := getWorldInstance().items
		sellableItems := make([]Item, 0)

		for n := 0; n < rand.Intn(len(allItemsPresent)); n++ {
			sellableItems = append(sellableItems, getWorldInstance().items[rand.Intn(len(allItemsPresent))])
		}

		writeToPlayer(player.conn, Reduce(sellableItems, func (item Item, acc string) string {
			return acc + item.description
		}, ""))

		var options = map[string]func(modifiers []string, items *[]Item){
			"buy": player.buyItem,
			"sell": player.sellItem,
		}

		for true {
			tmp := make([]byte, 256)
			player.conn.Read(tmp)

			// Parse input by changing all special
			parsedInput := strings.Split(nonAlphanumericRegex.ReplaceAllString(string(bytes.Trim(tmp, "\x00")), ""), " ")

			if ContainsKey(options, parsedInput[0]) {
				options[parsedInput[0]](parsedInput[1:len(parsedInput)], &sellableItems)
			} else {
				if parsedInput[0] == "leave" {
					break
				} else if parsedInput[0] == "help" {
					writeToPlayer(player.conn, "You are interacting with an NPC - listed below are the actions you can undertake")
					for _, option := range GetKeys(options) {
						writeToPlayerCompact(player.conn, option)
					}
				} else {
					writeToPlayer(player.conn, "Unknown command")
				}
			}
		}
	},
}

func initaliseGame() {
	// Pick random start point within the array
	var point = Point{WIDTH / 2, HEIGHT / 2, 0, 0, nil}

	lastDirection := pickPerpendicularRandomDirection("north")

	// Generate a number of walks to make an actual dungeon
	for i := 0; i < MAX_TUNNELS; i++ {
		/*
			Pick random direction to walk which is perpendicular to the last direction
			If last was right/left, new one must be up/down
		*/
		randomDirection := pickPerpendicularRandomDirection(lastDirection)

		// Calculate how long a tunnel will be
		tunnelLength := rand.Intn(MAX_TUNNEL_LENGTH)

		for j := 0; j < tunnelLength; j++ {
			getNewPoint(randomDirection, &point)

			// Tiles become 0 on default and 1 when walked on to generate dungeon rooms
			getWorldInstance().worldMap[point.x][point.y] = 1
		}

		// Set last direction
		lastDirection = randomDirection
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
	for i := 0; i < rand.Intn(5); i++ {
		randomPoint := findFreeLocationInDungeon()
		getWorldInstance().items = append(getWorldInstance().items, Item{"hotspot", *randomPoint, true, EventObject})
	}

	// Generate a number of NPC's which are able to sell items
	npcNames, err := os.ReadFile("data/npcNames.txt")
	if err != nil {
		panic(err)
	}

	for _, name := range strings.Split(string(npcNames), "\n") {
		getWorldInstance().items = append(getWorldInstance().items, Item{name, *findFreeLocationInDungeon(), true, EventObject})
	}
}

func handleConnection(conn net.Conn) {
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
		"map": player.printMap,
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
			writeToPlayer(player.conn, "Unknown command")
		}
	}
}

// Channels are pipes which connect concurrent goroutines in which values can be send between

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


