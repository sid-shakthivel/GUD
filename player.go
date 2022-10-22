package main

import (
	"math"
	"fmt"
	"sort"
	"strconv"
	"net"
	"unicode"
)

// Players serve as clients to the server which navigate around the world
type Player struct {
	coordinates Point
	inventory   []Item
	conn net.Conn
	name string
	actions map[string]func(modifiers []string)
}

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

/*
Signature `move {direction} {distance}`
Player coordinates are manipulated in a direction until they hit distance or a wall
*/
func (player Player) move(modifiers []string) {
	// Parse input correctly
	if len(modifiers) < 2 || !isInt(modifiers[1]) {
		player.displayError()
		return
	}

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
			player.displayError()
		}
	}

	writeToPlayer(player.conn, "- Your position is [" + strconv.Itoa(player.coordinates.x) + "," + strconv.Itoa(player.coordinates.y) + "]")
	writeToPlayer(player.conn, "The map is below")
}

func (player Player) isWithinBoundry(newCoordinate Point) bool {
	if newCoordinate.x > 20 || newCoordinate.y >= 40 || getWorldInstance().worldMap[newCoordinate.x][newCoordinate.y] == 0 {
		return false
	}
	return true
}

/*
Signature `scan {distance}`
Allows player to scan nearby to idenify items and teleportation in all directions for a unit
*/
func (player Player) scan(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError() }
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
	// Check for parameters
	if len(modifiers) < 1 { player.displayError() }

	// Check if item is within the eventObject dictionary and run functions
	eventObject[modifiers[0]](player, modifiers[1:len(modifiers)])
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
func (player Player) locate(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError() }

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
TODO: Possible to make some sort of method to combine these two
*/
func (player Player) pickup(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError() }

	// Search items array for item requested to get item of description at coordinate

	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		fmt.Println("Item not found")
	} else {
		player.inventory = append(player.inventory, getWorldInstance().items[itemIndex])
		getWorldInstance().items = RemoveAtIndex(getWorldInstance().items, itemIndex)
	}
}

/*
Signature `drop {item name}`
Remove item from inventory and place at a coordinate
*/
func (player Player) drop(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 1 { player.displayError() }

	// Search inventory for the item, remove it from inventory and append to items global
	itemIndex := Find(player.inventory, func (item Item) bool {
		return item.description == modifiers[0]
	})

	if itemIndex < 0 {
		fmt.Println("Item not found")
	} else {
		getWorldInstance().items = append(getWorldInstance().items, player.inventory[itemIndex])
		player.inventory = RemoveAtIndex(player.inventory, itemIndex)
	}
}

/*
Signature `combine {item1 name} {item2 name}
Combines item to solve puzzles (when included)
*/
func (player Player) combine(modifiers []string) {
	// Check for parameters
	if len(modifiers) < 2 { player.displayError() }

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
		player.inventory = append(player.inventory, Item {player.inventory[firstItemPosition].description + player.inventory[secondItemPosition].description, player.inventory[firstItemPosition].coordinates, true, Random })

		// Remove both items from players inventory
		RemoveAtIndex(player.inventory, firstItemPosition)
		RemoveAtIndex(player.inventory, secondItemPosition)

		// Possibly should check if items can be combined - make a random method for that maybe
	} else {
		panic("User does not possess both items")
	}
}

// Quit the game for a player (close the connection)
func (player Player) quit(modifiers []string) {
	writeToPlayer(player.conn, "Farewell " + player.name + "!")
	player.conn.Close()
}

func (player Player) help(modifiers []string) {
	writeToPlayer(player.conn, "Here lies the possible combinations once can enter")
	writeToPlayer(player.conn, "\nMove\nScan\nInvestigate\nLocate\nPickup\nDrop\nCombine")

//	for _, action := range GetKeys(player.actions) {
//		writeToPlayer(player.conn, "test " + action)
//	}
}
func (player Player) displayError() {
	player.conn.Write([]byte("\nUnknown or invalid command \n\n"))
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



