package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"net"
)

// Players serve as clients to the server which navigate around the world
type Player struct {
	coordinates *Point
	inventory   []Item
	conn net.Conn
	name string
	armour *Item
	weapon *Item
	health int
	actions map[string]func(modifiers []string)
}

func NewPlayer(coordinates *Point, conn net.Conn, name string) *Player {
	inventory := make([]Item, 1)
	inventory[0] = Item{ "blonde", Point {15, 20, 0, 0, nil}, true, Random}

	p := new(Player)
	p.coordinates = coordinates
	p.inventory = inventory
	p.conn = conn
	p.name = name
	p.health = 100

	return p
}

/*
Signature `move {direction} {distance}`
Player coordinates are manipulated in a direction until they hit distance or a wall
*/
func (player Player) move(modifiers []string) {
	// Parse input correctly

	if len(modifiers) < 2 || !isInt(modifiers[1]) || !Contains(GetKeys(directions), modifiers[0]) {
		player.displayError("")
		return
	}

	distance, err := strconv.Atoi(modifiers[1])
	if err != nil {
		panic(err)
	}

	for i := 0; i < distance; i++ {
		oldCoordinates := player.coordinates
		getNewPoint(modifiers[0], player.coordinates)

		if !player.isWithinPlayableRegion() {
			player.coordinates = oldCoordinates
			break
		}
	}

	writeToPlayer(player.conn, "- Your position is " + player.coordinates.format())
	writeToPlayer(player.conn, "The map is below")
	player.printMap()
}

func (player Player) isWithinPlayableRegion() bool {
	if (*player.coordinates).y > HEIGHT || (*player.coordinates).x > WIDTH || getWorldInstance().worldMap[(*player.coordinates).x][(*player.coordinates).y] == 1 {
		return false
	}
	return true
}

/*
Signature `scan {distance}`
Allows player to scan nearby to identify items and eventObjects in all directions
*/
func (player Player) scan(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	// Looks for unit block in all directions to check for item and reports back to user

	distance, err := strconv.Atoi(modifiers[0])
	if err != nil {
		panic(err)
	}

	// Check coordiantes of all items if they are within distance
	for _, item := range getWorldInstance().items {
		if int(math.Abs(float64(item.coordinates.x - player.coordinates.x))) <= distance && int(math.Abs(float64(item.coordinates.y - player.coordinates.y))) <= distance {
			writeToPlayer(player.conn, "Found: " + item.description + " at " + item.coordinates.format())
		}
	}
	writeToPlayer(player.conn, "Scan finished")
}

/*
Signature `investigate {item}`
Allows player to investigate items they encounter including hotspots
*/
func (player Player) investigate(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError("") }

	// Check if item is within the eventObject dictionary and run functions
	eventObject[modifiers[0]](player, modifiers[1:len(modifiers)])
}

// Check if a point slice contains a point with the same coordiantes
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
Uses A* path finding algorithm to work out the shortest path between the user and an item
Dispays the path to the user
*/
func (player Player) locate(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError("") }

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

		openNodes = append(openNodes, *player.coordinates) // Add starting node

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

				fmt.Println("Path uncovered")
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
}

/*
Signature `pickup {item name}`
Add item to inventory
*/
func (player Player) pickup(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError("") }

	// Search items array for item requested to get item of description at coordinate
	itemIndex := Find(getWorldInstance().items, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item requested is not present within map")
		return
	}

	item := getWorldInstance().items[itemIndex]

	if item.coordinates.x != player.coordinates.x || item.coordinates.y != player.coordinates.y {
		player.displayError("You are not at the location of the item")
		return
	}

	player.inventory = append(player.inventory, getWorldInstance().items[itemIndex])
	getWorldInstance().items = RemoveAtIndex(getWorldInstance().items, itemIndex)

	writeToPlayer(player.conn, "Picked up " + item.description)
}

/*
Signature `drop {item name}`
Remove item from inventory and place at a coordinate
*/
func (player Player) drop(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	// Search inventory for the item, remove it from inventory and append to items global
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item requested is not present within your inventory")
		return
	}

	getWorldInstance().items = append(getWorldInstance().items, player.inventory[itemIndex])
	player.inventory = RemoveAtIndex(player.inventory, itemIndex)
}

/*
Signature `combine {item1 name} {item2 name}
Combines item to solve puzzles (when included)
*/
func (player Player) combine(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 2 {
		player.displayError("")
		return
	}

	if modifiers[0] == modifiers[1] {
		player.displayError("You have chosen to combine the same item")
		return
	}

	firstItemPosition := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	secondItemPosition := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[1]
	})

	if firstItemPosition < 0 || secondItemPosition < 0 {
		player.displayError("You do not possess both items")
		return
	}

	// Add a new combined item to inventory
	player.inventory = append(player.inventory, Item {player.inventory[firstItemPosition].description + player.inventory[secondItemPosition].description, player.inventory[firstItemPosition].coordinates, true, Random })

	// Remove both items from players inventory
	RemoveAtIndex(player.inventory, firstItemPosition)
	RemoveAtIndex(player.inventory, secondItemPosition)
}

/*
Signature `equip {item1 name}`
Equips the weapon/armour to a player
*/
func (player Player) equip(modifiers[] string) {
	// Get item information
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item not found within inventory")
		return
	}

	item := getWorldInstance().items[itemIndex]

	switch item.itemType {
	case Armour:
		player.armour = &item
		writeToPlayer(player.conn, "Equiped " + item.description + " as armour")
	case Weapon:
		player.weapon = &item
		writeToPlayer(player.conn, "Equiped " + item.description + " as weapon")
	default:
		player.displayError("Item cannot be equiped")
	}
}

/*
Signature `unequip {item1 name}`
Equips the weapon/armour to a player
*/
func (player Player) unequip(modifiers[] string) {
	// Get item information
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item not found within inventory")
		return
	}

	item := getWorldInstance().items[itemIndex]

	switch item.itemType {
	case Armour:
		player.armour = nil
		writeToPlayer(player.conn, "Unequiped " + item.description + " as armour")
	case Weapon:
		player.weapon = nil
		writeToPlayer(player.conn, "Unequiped " + item.description + " as weapon")
	default:
		player.displayError("Item cannot be unequiped")
	}
}

// Quit the game for a player (close the connection)
func (player Player) quit(modifiers []string) {
	writeToPlayer(player.conn, "Farewell " + player.name + "!")
	player.conn.Close()
}

func (player Player) help(modifiers []string) {
	writeToPlayer(player.conn, "Here lies the possible combinations once can enter")
	writeToPlayer(player.conn, "Move\nScan\nInvestigate\nLocate\nPickup\nDrop\nCombine\nStats\nEquip\nUnequip\nQuit\nHelp")
}
func (player Player) viewStats(modifiers []string) {
	player.conn.Write([]byte("\nName : " + player.name + "\n"))
	player.conn.Write([]byte("Position: " + player.coordinates.format() + "\n"))
	if player.armour == nil {
		player.conn.Write([]byte("Armour: Not Equiped" + "\n"))
	} else {
		player.conn.Write([]byte("Armour: " + player.armour.description + "\n"))
	}

	if player.weapon == nil {
		player.conn.Write([]byte("Weapon: Not Equiped" + "\n"))
	} else {
		player.conn.Write([]byte("Weapon: " + player.weapon.description + "\n"))
	}

	player.conn.Write([]byte("Inventory contents: "))

	for _, item := range player.inventory {
		player.conn.Write([]byte(item.description))
	}
	player.conn.Write([]byte("\n\n"))
}


func (player Player) printMap() {
	worldMap := getWorldInstance().worldMap
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			if i == player.coordinates.x && j == player.coordinates.y {
				player.conn.Write([]byte("X"))
			} else if worldMap[i][j] == 1 {
				player.conn.Write([]byte("#"))
			} else {
				player.conn.Write([]byte("/"))
			}
		}
		player.conn.Write([]byte("\n"))
	}
	player.conn.Write([]byte("\n"))
}

func (player Player) displayError(message string) {
	if message == "" {
		player.conn.Write([]byte("\nUnknown or invalid command \n\n"))
	} else {
		player.conn.Write([]byte("\n" + message + "\n\n"))
	}
}



