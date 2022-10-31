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
	player := NewPlayer(NewPoint(15, 10), conn, "Example", getWorldInstance().towns[0])

	player.write(BANNER)

	// Create new player and retrieve name
	player.write("Good day fellow union member!")
	player.write("By what do you wish to be addressed by?")

	nameBytes := make([]byte, 256)
	_, _ = conn.Read(nameBytes)

	nameStr := nonAlphanumericRegex.ReplaceAllString(string(nameBytes), "")

	player.name = nameStr

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
		"jump":    player.jump,
		"eat": 	   player.eat,
		"help":    player.help,
	}

	player.actions = actions

	player.write("Welcome to GUD! "+nameStr)

	player.write(player.currentTown.description)
	player.listRoutes()

	for {
		// Parse commands a user enters
		tmp := make([]byte, 256)
		conn.Read(tmp)

		// Parse input by changing all special
		parsedInput := strings.Split(nonAlphanumericRegex.ReplaceAllString(string(bytes.Trim(tmp, "\x00")), ""), " ")

		for _, word := range parsedInput {
			word = strings.ToLower(word)
		}

		if ContainsKey(actions, parsedInput[0]) {
			actions[parsedInput[0]](parsedInput[1:len(parsedInput)])
		} else {
			player.write("Unknown command")
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
