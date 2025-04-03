package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type Room struct {
	Name  string
	Links []string
}

type AntFarm struct {
	Ants  int
	Rooms map[string]*Room
	Start string
	End   string
}

func ParseFile(file *os.File) *AntFarm {
	scanner := bufio.NewScanner(file)
	rooms := make(map[string]*Room)
	flags := make(map[string]bool)
	var fileContent []string

	startRoom := ""
	endRoom := ""
	roomProcessingDone := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "##") {
			continue
		}
		if line == "" {
			continue
		}
		fileContent = append(fileContent, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("ERROR: reading file: %v", err)
	}

	if len(fileContent) == 0 {
		log.Fatal("ERROR: empty file")
	}

	antsNumber, err := strconv.Atoi(fileContent[0])
	if err != nil {
		log.Fatal("ERROR: invalid data format, first line must be number of ants")
	}
	if antsNumber < 1 {
		log.Fatal("ERROR: invalid number of ants")
	}

	// Parse rooms and tunnels
	for i := 1; i < len(fileContent); i++ {
		line := fileContent[i]

		// Handle start room
		if line == "##start" {
			if flags["##start"] {
				log.Fatal("ERROR: multiple start rooms defined")
			}
			flags["##start"] = true
			if i+1 >= len(fileContent) {
				log.Fatal("ERROR: no room defined after ##start")
			}
			i++
			startRoom = validateRoom(fileContent[i], rooms)
			continue
		}

		// Handle end room
		if line == "##end" {
			if flags["##end"] {
				log.Fatal("ERROR: multiple end rooms defined")
			}
			flags["##end"] = true
			if i+1 >= len(fileContent) {
				log.Fatal("ERROR: no room defined after ##end")
			}
			i++
			endRoom = validateRoom(fileContent[i], rooms)
			continue
		}

		// Check if the line is a tunnel definition (contains a "-")
		if strings.Contains(line, "-") {
			roomProcessingDone = true
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				log.Fatalf("ERROR: invalid tunnel format: %s", line)
			}

			room1, room2 := parts[0], parts[1]
			if _, exists1 := rooms[room1]; !exists1 {
				log.Fatalf("ERROR: room %s does not exist", room1)
			}
			if _, exists2 := rooms[room2]; !exists2 {
				log.Fatalf("ERROR: room %s does not exist", room2)
			}

			if room1 == room2 {
				log.Fatalf("ERROR: room cannot link to itself: %s", line)
			}

			// Check for duplicate links
			for _, link := range rooms[room1].Links {
				if link == room2 {
					log.Fatalf("ERROR: duplicate tunnel: %s", line)
				}
			}

			rooms[room1].Links = append(rooms[room1].Links, room2)
			rooms[room2].Links = append(rooms[room2].Links, room1)
			continue
		}

		// If not a tunnel and we've already started processing tunnels we continue to the next line
		if roomProcessingDone {
			continue
		}

		// Must be a room definition if we get here
		validateRoom(line, rooms)
	}

	if !flags["##start"] {
		log.Fatal("ERROR: no start room found")
	}
	if !flags["##end"] {
		log.Fatal("ERROR: no end room found")
	}
	if len(rooms) < 2 {
		log.Fatal("ERROR: insufficient number of rooms")
	}

	farm := &AntFarm{
		Ants:  antsNumber,
		Rooms: rooms,
		Start: startRoom,
		End:   endRoom,
	}

	return farm
}

func validateRoom(line string, rooms map[string]*Room) string {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		log.Fatalf("ERROR: invalid room format: %s", line)
	}

	roomName := parts[0]
	if roomName == "" || strings.HasPrefix(roomName, "L") || strings.HasPrefix(roomName, "#") {
		log.Fatalf("ERROR: invalid room name: %s", roomName)
	}

	// Check for duplicate room names
	if _, exists := rooms[roomName]; exists {
		log.Fatalf("ERROR: duplicate room name: %s", roomName)
	}

	rooms[roomName] = &Room{
		Name:  roomName,
		Links: []string{},
	}

	return roomName
}
