package main

/*
Items are objects which are located around the map
They can be obtained/dropped by a player
*/
type Item struct {
	description string
	coordinates Point
	isActive    bool
	itemType ItemType
}

type ItemType int

// Type of item is used to define the actions that can be taken with an item
const (
	Armour ItemType = iota
	Weapon
	Random
	EventObject
)