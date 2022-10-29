package main

import (
	"math/rand"
	"os"
	"strings"
	"sync"
	"fmt"
)

// World simply consists of all of the rooms put together
type World struct {
	towns []Town // Slice of towns
}

/*
Singleton restricts instaniation of struct to a single instance
Singletons also provide a global access to an instance and protects it from being overwritten
*/
var once sync.Once
var worldInstance *World

func getWorldInstance() *World {
	if worldInstance == nil {
		once.Do(
			func() {
				worldInstance = NewWorld()
			})
	}

	return worldInstance
}

func NewWorld() *World {
	w := new(World)

	townNamesFile, _ := os.ReadFile("data/towns.txt")
	townNames := strings.Split(string(townNamesFile), "\n")
	townName := "Unknown"

	var towns []Town

	for i := 1; i < rand.Intn(5) + 3; i++ {
		// Create new room and add it to slice
		townNames, townName = GetRandomAndRemove(townNames)
		newTown := NewTown(townName)
		towns = append(towns, *newTown)

		// Pick room and make the new room adjacent to it - check it's not the newly created room
		if (i - 1) > 0 {
			// Need to go back out
			towns[i - 2].adjacentTowns[0] = *newTown
		}
	}

	w.towns = towns
	fmt.Println("Created", len(towns), "rooms")

	return w
}

/*
Town consists of a general dungeon layout, items, events, and adjacent rooms
*/
type Town struct {
	dungeonLayout [WIDTH][HEIGHT]int // Pixel array of map
	items         []Item             // Slice of items
	events        []Event            // Slice of events
	name          string             // Name
	adjacentTowns []Town             // Adjoining to this room in a specific direction [North, South, East, West]
	description string
}

func NewTown(name string) *Town {
	r := new(Town)

	r.name = name
	r.adjacentTowns = make([]Town, 4)

	// Pick random description
	descriptions, _ := os.ReadFile("data/townDescription.txt")
	townDescriptions := strings.Split(string(descriptions), "\n")
	r.description = strings.Replace(townDescriptions[rand.Intn(len(townDescriptions))], "{}", r.name, -1)

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
			r.dungeonLayout[point.x][point.y] = 1
		}

		// Set last direction
		lastDirection = randomDirection
	}

	// Retrieve items which are predefined in a text file and add them into world
	content, _ := os.ReadFile("data/items.txt")

	itemNames := strings.Split(string(content), "\n")

	for i := 0; i < rand.Intn(len(itemNames)); i++ {
		// Find a random location within the dungeon to place the item which is free
		randomPoint := findFreeLocationInDungeon(r.dungeonLayout)
		itemType := Random

		itemName := itemNames[i]

		// Check type
		if strings.Contains(itemName, "armour") {
			itemType = Armour
		} else if strings.Contains(itemName, "sword") || strings.Contains(itemName, "spear") {
			itemType = Weapon
		}

		// Push new item to global items array
		r.items = append(r.items, Item{itemName, *randomPoint, true, itemType})
	}

	// Generate a random number of hotspots to be placed inside the world
	for i := 0; i < rand.Intn(5); i++ {
		r.events = append(r.events, Event{*findFreeLocationInDungeon(r.dungeonLayout), Hotspot, ""})
	}

	// Generate NPC's from data files
	npcNames, _ := os.ReadFile("data/npcNames.txt")

	for _, name := range strings.Split(string(npcNames), "\n") {
		r.events = append(r.events, Event{*findFreeLocationInDungeon(r.dungeonLayout), NPC, name})
	}

	// Generate enemies from data files
	enemyNames, _ := os.ReadFile("data/enemies.txt")

	for _, name := range strings.Split(string(enemyNames), "\n") {
		r.events = append(r.events, Event{*findFreeLocationInDungeon(r.dungeonLayout), Enemy, name})
	}

	return r
}

func (town *Town) checkAdjacentTown(direction string) (bool, string, int) {
	switch direction {
	case "north":
		return town.checkEmptyTown(0)
	case "south":
		return town.checkEmptyTown(1)
	case "east":
		return town.checkEmptyTown(2)
	case "west":
		return town.checkEmptyTown(3)
	default:
		return false, "Unknown dirction", 0
	}
}

func (town *Town) checkEmptyTown(index int) (bool, string, int) {
	if town.adjacentTowns[index].name == ""  {
		return false, "No town that way m8", index
	} else {
		return true, "", 0
	}
}

func convertToText(index int) string {
	switch index {
	case 0:
		return "north"
	case 1:
		return "south"
	case 2:
		return "east"
	case 3:
		return "west"
	default:
		return "unknown"
	}
}