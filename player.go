package main

import (
	"math"
	"fmt"
	"sort"
	"strconv"
	"net"
	"math/rand"
)

// Player serve as clients to the server which navigate around the world
type Player struct {
	coordinates *Point
	inventory   []Item
	conn net.Conn
	name string
	armour *Item
	weapon *Item
	health int
	gold int
	actions map[string]func(modifiers []string)
	currentTown Town
}

func NewPlayer(coordinates *Point, conn net.Conn, name string, town Town) *Player {
	inventory := make([]Item, 1)
	inventory[0] = Item{ "blonde", Point {15, 20, 0, 0, nil}, true, Random}

	p := new(Player)
	p.coordinates = coordinates
	p.inventory = inventory
	p.conn = conn
	p.name = name
	p.health = 100
	p.gold = 100
	p.currentTown = town

	return p
}

/*
Signature `move {direction} {distance}`
Player coordinates are manipulated in a direction within an individual town until they hit distance or a wall
*/
func (player *Player) move(modifiers []string) {
	// Parse input correctly

	if len(modifiers) < 2 || !isInt(modifiers[1]) || !Contains(GetKeys(directions), modifiers[0]) {
		player.displayError("")
		return
	}

	distance, err := strconv.Atoi(modifiers[1])
	if err != nil {
		player.displayError("")
		return
	}

	for i := 0; i < distance; i++ {
		oldX := (*player.coordinates).x
		oldY := (*player.coordinates).y
		getNewPoint(modifiers[0], player.coordinates)

		if !player.isWithinPlayableRegion() {
			player.coordinates = NewPoint(oldX, oldY)
			player.write("A wall blocks your path - one must circumvent it")
			break
		}
	}

	player.write("- Your position is " + player.coordinates.format())

	// Check if player has encountered an event and trigger one
	for _, event := range player.currentTown.events {
		if event.coordinates.x == player.coordinates.x && event.coordinates.y == player.coordinates.y {
			player.write("You have found a " + event.eventType.String())
			events[event.eventType](player, event)
		}
	}
}

func (player *Player) isWithinPlayableRegion() bool {
	if (*player.coordinates).y > HEIGHT || (*player.coordinates).x > WIDTH || player.currentTown.dungeonLayout[(*player.coordinates).x][(*player.coordinates).y] == 0 {
		return false
	}
	return true
}

/*
Signature `scan {distance}`
Allows player to scan nearby to identify items and eventObjects in all directions
*/
func (player *Player) scan(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 || !isInt(modifiers[0]) {
		player.displayError("")
		return
	}

	// Looks for unit block in all directions to check for item and reports back to user
	distance, err := strconv.Atoi(modifiers[0])
	if err != nil {
		player.displayError("")
		return
	}

	player.writeCompact("Items:")

	// Check coordiantes of all items if they are within distance
	for _, item := range player.currentTown.items {
		if int(math.Abs(float64(item.coordinates.x - player.coordinates.x))) <= distance && int(math.Abs(float64(item.coordinates.y - player.coordinates.y))) <= distance {
			player.writeCompact("Found: " + item.description + " at " + item.coordinates.format())
		}
	}
	
	player.writeCompact("Events:")

	for _, event := range player.currentTown.events {
		if int(math.Abs(float64(event.coordinates.x - player.coordinates.x))) <= distance && int(math.Abs(float64(event.coordinates.y - player.coordinates.y))) <= distance {
			player.writeCompact("Found: " + event.eventType.String() + " at " + event.coordinates.format())
		}
	}

	player.write("Scan finished")
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
func (player *Player) locate(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError("") }

	// Find item position in world
	itemIndex, item := Find(player.currentTown.items, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item is not within the world")
		return
	}

	itemPosition := item.coordinates

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

		// Check if target is found
		if currentNode.x == itemPosition.x && currentNode.y == itemPosition.y {
			path := make([]Point, 0)

			node := currentNode
			path = append(path, node)

			player.write("A path has been uncovered - follow it to find the " + player.currentTown.items[itemIndex].description)

			player.write(node.format())

			for node.parent != nil {
				node = *node.parent
				path = append(path, node)
				player.writeCompact(node.format())
			}

			player.writeCompact("")

			return
		} else {
			// Create a list of adjacent nodes which are walkable from the current node and not closed

			neighbour1 := *NewPoint(min(currentNode.x + 1, WIDTH - 1), currentNode.y)
			neighbour2 := *NewPoint(max(currentNode.x - 1, 0), currentNode.y,)
			neighbour3 := *NewPoint(currentNode.x, min(currentNode.y + 1, HEIGHT - 1))
			neighbour4 := *NewPoint(currentNode.x, max(currentNode.y - 1, 0))
			neighbour5 := *NewPoint(min(currentNode.x + 1, WIDTH - 1), min(currentNode.y + 1, HEIGHT - 1))
			neighbour6 := *NewPoint(min(currentNode.x + 1, WIDTH -1), max(currentNode.y - 1, 0))
			neighbour7 := *NewPoint(max(currentNode.x - 1, 0), min(currentNode.y + 1, HEIGHT - 1))
			neighbour8 := *NewPoint(max(currentNode.x - 1, 0), max(currentNode.y - 1, 0))

			neighbours := [8]Point { neighbour1, neighbour2, neighbour3, neighbour4, neighbour5, neighbour6, neighbour7, neighbour8 }

			for _, neighbour := range neighbours {
				// Check if it's walkable (world[neighbour.x][neighbour.y] == 1) and not on the closed list
				if !neighbour.ContainsPoint(closedNodes) && player.currentTown.dungeonLayout[neighbour.x][neighbour.y] == 1 {
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

	player.displayError("Cannot locate item")
}

/*
Signature `pickup {item}`
Add item to inventory
*/
func (player *Player) pickup(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError("") }

	// Search items array for item requested to get item
	itemIndex, item := Find(player.currentTown.items, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item requested is not present within map")
		return
	}

	if item.coordinates.x != player.coordinates.x || item.coordinates.y != player.coordinates.y {
		player.displayError("You are not at the location of the item")
		return
	}

	// Check the type of the item to determine what to do

	switch item.itemType {
	case Random, Weapon, Armour:
		player.inventory = append(player.inventory, item)
		player.write("Picked up " + item.description)
		player.currentTown.items = RemoveAtIndex(player.currentTown.items, itemIndex)
	default:
		player.displayError("Cannot pickup an events object - investigate it pronto")
		return
	}
}

/*
Signature `drop {item}`
Remove item from inventory and place at a coordinate
*/
func (player *Player) drop(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	// Search inventory for the item, remove it from inventory and append to items global
	itemIndex, item := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item requested is not present within your inventory")
		return
	}

	// Remove from weapon/armour
	if player.weapon != nil && player.weapon.description == item.description {
		player.weapon = nil
	}

	if player.armour != nil && player.armour.description == item.description {
		player.armour = nil
	}

	player.currentTown.items = append(player.currentTown.items, item)
	player.inventory = RemoveAtIndex(player.inventory, itemIndex)

	player.write("Dropped " + item.description)
}

/*
Signature `combine {item1} {item2}
Combines item to solve puzzles (when included)
*/
func (player *Player) combine(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 2 {
		player.displayError("")
		return
	}

	if modifiers[0] == modifiers[1] {
		player.displayError("You have chosen to combine the same item")
		return
	}

	firstItemPosition, firstItem := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	secondItemPosition, secondItem := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[1]
	})

	if firstItemPosition < 0 || secondItemPosition < 0 {
		player.displayError("You do not possess both items")
		return
	}

	if firstItem.itemType != Random || secondItem.itemType != Random {
		player.displayError("You can only combine random items")
		return
	}

	combinedItem := Item {firstItem.description + secondItem.description, firstItem.coordinates, true, Random }

	player.inventory = append(player.inventory, combinedItem)

	// Remove both items from players inventory
	player.inventory = RemoveAtIndex(player.inventory, firstItemPosition)
	player.inventory = RemoveAtIndex(player.inventory, secondItemPosition)

	player.write("Combined " + modifiers[0] + " and " + modifiers[1] + " to create a " + combinedItem.description)
}

/*
Signature `equip {item1}`
Equips the weapon/armour to a player
*/
func (player *Player) equip(modifiers[] string) {
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	// Get item information
	itemIndex, item := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item not found within inventory")
		return
	}

	switch item.itemType {
	case Armour:
		player.armour = &item
		player.write("Equiped " + item.description + " as armour")
	case Weapon:
		player.weapon = &item
		player.write("Equiped " + item.description + " as weapon")
	default:
		player.displayError("Item cannot be equiped")
	}
}

/*
Signature `unequip {item1}`
Equips the weapon/armour to a player
*/
func (player *Player) unequip(modifiers[] string) {
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	// Get item information
	itemIndex, item := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Item not found within inventory")
		return
	}

	switch item.itemType {
	case Armour:
		if player.armour == nil {
			player.displayError("Armour is not equipped")
			return
		}

		if player.armour.description != item.description {
			player.displayError(item.description + " is not equipped fool")
			return
		}

		player.armour = nil
		player.write("Unequiped " + item.description + " as armour")
	case Weapon:
		if player.weapon == nil {
			player.displayError("Weapon is not equipped")
			return
		}

		if player.weapon.description != item.description {
			player.displayError(item.description + " is not equipped fool")
			return
		}

		player.weapon = nil
		player.write("Unequiped " + item.description + " as weapon")
	default:
		player.displayError("Item cannot be unequiped")
	}
}

// Quit the game for a player (close the connection)
func (player *Player) quit(modifiers []string) {
	player.write("Farewell " + player.name + "!")
	player.conn.Close()
}

func (player *Player) help(modifiers []string) {
	player.write("Here lies the possible combinations once can enter")

	for _, action := range GetKeys(player.actions) {
		player.writeCompact(action)
	}
	player.writeCompact("")
}
func (player *Player) viewStats(modifiers []string) {
	player.conn.Write([]byte("\nName : " + player.name + "\n"))
	player.conn.Write([]byte("Position: " + player.coordinates.format() + "\n"))
	player.conn.Write([]byte("Health: " + strconv.Itoa(player.health) + "\n"))
	player.conn.Write([]byte("Gold coins: " + strconv.Itoa(player.gold) + "\n"))

	if player.armour == nil {
		player.writeCompact("Armour: Not Equiped")
	} else {
		player.writeCompact("Armour: " + player.armour.description)
	}

	if player.weapon == nil {
		player.writeCompact("Weapon: Not Equiped")
	} else {
		player.writeCompact("Weapon: " + player.weapon.description)
	}

	player.conn.Write([]byte("Inventory contents: "))

	for _, item := range player.inventory {
		player.conn.Write([]byte(item.description + " "))
	}
	player.write("")
}


func (player *Player) printMap(modifiers []string) {
	player.writeCompact("")

	worldMap := player.currentTown.dungeonLayout
	for i := 0; i < HEIGHT; i++ {
		for j := 0; j < WIDTH; j++ {
			if j == player.coordinates.x && i == player.coordinates.y {
				player.conn.Write([]byte("X"))
			} else if worldMap[j][i] == 1 {
				player.conn.Write([]byte("#"))
			} else {
				player.conn.Write([]byte("/"))
			}
		}
		player.writeCompact("")
	}
	player.writeCompact("")
}

/*
Signature: `buy {itemName}`
Allows player to buy an item from a seler in exchange for gold
*/
func (player *Player) buyItem(modifiers []string, items *[]Item) {
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	itemIndex, item := Find(*items, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("Sorry I do not sell that item")
		return
	}

	price := rand.Intn(10) // Generate random price

	if player.gold - price < 0 {
		player.displayError("You possess insufficient funds to purchase from the vendor")
		return
	}

	player.inventory = append(player.inventory, item)
	player.write("Purchased " + item.description + " for " + strconv.Itoa(price) + " gold")
	player.gold -= price
	*items = RemoveAtIndex(*items, itemIndex)
}

/*
Signature: `sell {itemName}`
Allows player to sel an item they possess in exchange for gold
*/
func (player *Player) sellItem(modifiers []string, items *[]Item) {
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	itemIndex, item := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		player.displayError("You do not possess such an item")
		return
	}

	price := rand.Intn(10) // Generate random price

	*items = append(*items, item)
	player.write("Sold " + item.description + " for " + strconv.Itoa(price) + " gold")
	player.gold += price
	player.inventory = RemoveAtIndex(player.inventory, itemIndex)
}

/*
Signature: `jump {direction}`
Transports a player to another town within the world
*/
func (player *Player) jump(modifiers[]string) {
	if len(modifiers) > 1 {
		player.displayError("")
		return
	}

	isRoom, message, townIndex := player.currentTown.checkAdjacentTown(modifiers[0])

	if !isRoom {
		player.displayError(message)
		return
	}

	// Move player and provide a random town description
	player.currentTown = player.currentTown.adjacentTowns[townIndex]

	player.writeCompact("")
	player.write(player.currentTown.description)
	player.listRoutes()

	fmt.Println("test")
}

/*
Signature: `eat {item}
Allows user to increase health (by random amount) by eating food within inventory
*/
func (player *Player) eat(modifiers []string) {
	if len(modifiers) < 1 {
		player.displayError("")
		return
	}

	itemIndex, item := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < -1 {
		player.displayError("Item is not within index")
		return
	}

	if item.itemType != Food {
		player.displayError("You can't eat that")
		return
	}

	player.inventory = RemoveAtIndex(player.inventory, itemIndex)
	player.health = min(rand.Intn(10) + player.health, 100)

	player.write("Your health is now: " + strconv.Itoa(player.health))
}

func (player *Player) listRoutes() {
	for _, route := range player.currentTown.getRoutes() {
		player.writeCompact(route)
	}
	player.writeCompact("")
}

func (player *Player) write(text string) {
	player.conn.Write([]byte(text + "\n\n"))
}

func (player *Player) writeCompact(text string) {
	player.conn.Write([]byte(text + "\n"))
}

func (player *Player) displayError(message string) {
	if message == "" {
		player.write("\nInvalid command")
	} else {
		player.write("\n" + message)
	}
}