package main

import (
	"bytes"
	"math/rand"
	"os"
	"strings"
)

/*
Events are similar to items however cannot be obtained
Events interact with player and transform the world
*/
type Event struct {
	coordinates Point
	eventType   EventType
	name        string ""
}

type EventType int

const (
	Hotspot EventType = iota
	NPC
	Enemy
)

func (eventType EventType) String() string {
	switch eventType {
	case Hotspot:
		return "hotspot"
	case NPC:
		return "humanoid"
	case Enemy:
		return "a dreaded foe"
	default:
		return "unknown"
	}
}

// Dictionary called events of strings to functions which transmute the player eventObjects
var events = map[EventType]func(player *Player, event Event){
	Hotspot: func(player *Player, event Event) {
		// Finding a hotspot moves the player to a random location within the dungeon
		// player.coordinates = findFreeLocationInDungeon()

		player.write("You have been deported to "+player.coordinates.format())

		// Destroy event
		eventIndex, event := Find(player.currentTown.events, func(innerEvent Event) bool {
			return innerEvent == event
		})

		player.currentTown.events = RemoveAtIndex(player.currentTown.events, eventIndex)
	},
	NPC: func(player *Player, event Event) {
		// NPC's sell random items to user on their request
		player.write("Goodday fellow union member I am "+event.name+"! What would you like to buy?")
		player.write("I sell the following items: ")

		allItemsPresent := player.currentTown.items
		sellableItems := make([]Item, 0)

		for n := 0; n < rand.Intn(len(allItemsPresent)); n++ {
			item := player.currentTown.items[rand.Intn(len(allItemsPresent))]
			sellableItems = append(sellableItems, item)
			player.conn.Write([]byte(item.description + ","))
		}
		player.writeCompact("\n")

		var options = map[string]func(modifiers []string, items *[]Item){
			"buy":  player.buyItem,
			"sell": player.sellItem,
		}

		for {
			tmp := make([]byte, 256)
			player.conn.Read(tmp)

			// Parse input by changing all special characters
			parsedInput := strings.Split(nonAlphanumericRegex.ReplaceAllString(string(bytes.Trim(tmp, "\x00")), ""), " ")

			if ContainsKey(options, parsedInput[0]) {
				options[parsedInput[0]](parsedInput[1:len(parsedInput)], &sellableItems)
			} else {
				if parsedInput[0] == "leave" {
					player.write("Good bye for now!")
					break
				} else if parsedInput[0] == "help" {
					player.write("You are interacting with an NPC - listed below are the actions you can undertake")
					for _, option := range GetKeys(options) {
						player.writeCompact(option)
					}
					player.writeCompact("leave")
					player.write("help")
				} else {
					player.write("Unknown command")
				}
			}
		}
	},
	Enemy: func(player *Player, event Event) {
		// Get appropriate data
		minimalAttacks, _ := os.ReadFile("data/attack/minimal.txt")
		moderateAttacks, _ := os.ReadFile("data/attack/moderate.txt")
		majorAttacks, _ := os.ReadFile("data/attack/major.txt")
		minimalResponse, _ := os.ReadFile("data/attackResponse/minimal.txt")
		moderateResponse, _ := os.ReadFile("data/attackResponse/moderate.txt")
		majorResponse, _ := os.ReadFile("data/attackResponse/major.txt")

		attacks := [][]string{strings.Split(string(minimalAttacks), "\n"), strings.Split(string(moderateAttacks), "\n"), strings.Split(string(majorAttacks), "\n")}
		attackResponse := [][]string{strings.Split(string(minimalResponse), "\n"), strings.Split(string(moderateResponse), "\n"), strings.Split(string(majorResponse), "\n")}

		enemyDamage := 0
		playerDamage := 0

		// Battle commences for random duration
		for i := 0; i < rand.Intn(3); i++ {
			if player.weapon != nil {
				// Select player attack which is based upon minimal/moderate/major
				level := rand.Intn(len(attacks) - 1)
				player.writeCompact("Player "+attacks[level][rand.Intn(len(attacks)-1)])
				enemyDamage += level
			} else {
				// Select player attack which is based upon minimal/moderate
				level := rand.Intn(len(attacks) - 2)
				player.writeCompact("Player "+attacks[level][rand.Intn(len(attacks)-1)])
				enemyDamage += level
			}

			// Select enemy response which is based upon minimal/moderate
			player.writeCompact(attackResponse[rand.Intn(len(attacks)-2)][rand.Intn(len(attacks)-1)])

			// Select enemy attack which is based upon minimal/moderate/major
			level := rand.Intn(len(attacks) - 1)
			player.writeCompact(event.name+" "+attacks[level][rand.Intn(len(attacks)-1)])
			playerDamage += level

			// Select player response which is based upon minimal/moderate
			player.writeCompact(attackResponse[rand.Intn(len(attacks)-2)][rand.Intn(len(attacks)-1)])
		}

		// Pick winner depedning on the number of attacks and select a final major attack and major response
		if enemyDamage > playerDamage {
			player.writeCompact("Player "+attacks[2][rand.Intn(len(attacks)-1)])
			player.writeCompact(attackResponse[rand.Intn(len(attacks)-1)][rand.Intn(2-1)+1])
		} else {
			player.writeCompact(event.name+" "+attacks[2][rand.Intn(len(attacks)-1)])
			player.writeCompact(attackResponse[rand.Intn(len(attacks)-1)][rand.Intn(2-1)+1])
		}

		player.writeCompact("")

		player.health -= playerDamage
	},
}
