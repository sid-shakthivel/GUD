package main

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strings"
)

const WIDTH = 30             // Width of map
const HEIGHT = 15            // Height of map
const MAX_TUNNELS = 50       // Greatest number of turns algorithm can make
const MAX_TUNNEL_LENGTH = 30 // Greatest length of each tunnel the algorithm will choose before making a turn

const BANNER = `
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func handleConnection(conn net.Conn) {
	writeToPlayer(conn, BANNER)

	// Create new player and retrieve name
	//	writeToPlayer(conn, "Good day fellow union member!")
	//	writeToPlayer(conn, "By what do you wish to be addressed by?")
	//
	//	nameBytes := make([]byte, 256)
	//	_, _ = conn.Read(nameBytes)

	//	nameStr := nonAlphanumericRegex.ReplaceAllString(string(nameBytes), "")

	nameStr := "sid"

	inventory := make([]Item, 1)
	inventory[0] = Item{"blonde", Point{15, 8, 0, 0, nil}, true, Random}

	player := NewPlayer(NewPoint(15, 10), conn, nameStr, getWorldInstance().towns[0])

	// Dictionary of actions which players can undertake
	var actions = map[string]func(modifiers []string){
		"move":    player.move,
		"scan":    player.scan,
		"locate":  player.locate,
		"pickup":  player.pickup,
		"drop":    player.drop,
		"combine": player.combine,
		"stats":   player.viewStats,
		"equip":   player.equip,
		"unequip": player.unequip,
		"quit":    player.quit,
		"map":     player.printMap,
		"help":    player.help,
	}

	player.actions = actions

	writeToPlayer(player.conn, "Welcome to GUD! "+nameStr)

	writeToPlayer(player.conn, player.currentTown.description)

	for {
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
	getWorldInstance()
	fmt.Println("Game loaded")
	startServer()
}
