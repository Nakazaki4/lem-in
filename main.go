package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fileName := os.Args[1]
	file, _ := os.Open(fileName)
	farm := ParseFile(file)
	planAntsJourney(farm)
	printPerformance()
}

func planAntsJourney(farm *AntFarm) {
	// Get all possible paths first
	initialPaths := make([][]string, 0)
	for _, neighbor := range farm.Rooms[farm.Start].Links {
		initialPaths = append(initialPaths, findShortestPath(farm, neighbor, farm.End))
	}
	// This will hold all combinations of compatible paths
	allGroups := [][][]string{}

	// Try each path as a starting point
	for _, path := range initialPaths {
		if len(path) != 0 {
			// Start a new combination with this path
			group := [][]string{path}
			allGroups = append(allGroups, group)

			newFarm := rebuildGraph(copyGraph(farm), path)
			// Find all compatible paths recursively
			findCompatiblePaths(newFarm, &group, &initialPaths, &allGroups)
		}
	}
	// // Map to store evaluation metrics for each combination
	stepTurns := make(map[int][]int)

	for i, group := range allGroups {
		antsPerPath := antDistribution(farm.Ants, &group)

		// Calculate the turns using the first path's length and ants
		firstPathLength := len(group[0])
		firstPathAnts := antsPerPath[0]
		// this calculates turns
		totalTurns := firstPathLength - 1 + firstPathAnts

		totalSteps := 0
		// this calculates steps
		for j, path := range group {
			totalSteps += antsPerPath[j] * len(path)
		}

		stepTurns[i] = []int{totalSteps, totalTurns}
	}

	// Find the best combination using MinSteps
	bestGroupIndex := MinSteps(stepTurns)
	bestGroup := allGroups[bestGroupIndex]
	bestDistribution := antDistribution(farm.Ants, &bestGroup)

	simulateMovement(farm.End, farm.Ants, &bestDistribution, bestGroup)
}

func MinSteps(stepTurns map[int][]int) int {
	first := true
	minSteps, minTurns := 0, 0
	var index int
	for i, turns := range stepTurns {
		if first {
			minSteps = turns[0]
			minTurns = turns[1]
			first = false
			index = i
		} else if turns[1] < minTurns || (turns[1] == minTurns && turns[0] < minSteps) {
			minSteps = turns[0]
			minTurns = turns[1]
			index = i
		}
	}
	return index
}

func findCompatiblePaths(modifiedFarm *AntFarm, currentGroup *[][]string, allPaths *[][]string, allGroups *[][][]string) {
	possiblePaths := getAllPossiblePathsBfs(modifiedFarm)
	if len(possiblePaths) == 0 {
		return
	}

	for _, path := range possiblePaths {
		if len(path) == 0 {
			continue
		}
		// Add this path to our combination
		newGroup := make([][]string, len(*currentGroup))
		copy(newGroup, *currentGroup)
		newGroup = append(newGroup, path)
		// we should check if the path we're about to append doesn't intersect with any paths in the previous combination
		p := path[:len(path)-1]
		if !isCompatibleWithComb(currentGroup, &p) || len(newGroup) > modifiedFarm.Ants || len(newGroup) <= 0 {
			continue
		}
		*allGroups = append(*allGroups, newGroup)

		// Further modify the farm and continue recursively
		newFarm := rebuildGraph(copyGraph(modifiedFarm), path)
		findCompatiblePaths(newFarm, &newGroup, allPaths, allGroups)
	}
}

func isCompatibleWithComb(combination *[][]string, pathToAppend *[]string) bool {
	roomSet := make(map[string]struct{})
	for _, path := range *combination {
		for _, room := range path {
			roomSet[room] = struct{}{}
		}
	}

	for _, node := range *pathToAppend {
		if _, exists := roomSet[node]; exists {
			return false
		}
	}

	for n := range roomSet {
		delete(roomSet, n)
	}
	return true
}

func simulateMovement(end string, totalAnts int, antsPerPath *[]int, paths [][]string) int {
	ants := make([]*Ant, totalAnts)
	antID := 1

	for pathID, count := range *antsPerPath {
		for i := range count {
			ants[antID-1] = &Ant{
				ID:       antID,
				PathID:   pathID,
				Position: -i,
				Path:     paths[pathID],
				Finished: false,
			}
			antID++
		}
	}

	finished := 0
	counter := 0

	for finished < totalAnts {
		counter++
		// Track movements for this turn
		moves := make(map[string]int) // room -> antID
		movementsMade := false
		roundMoves := []string{} // Collect all moves for this round

		// Move each ant
		for _, ant := range ants {
			if ant.Finished {
				continue
			}

			// Check if ant should start moving
			if ant.Position < 0 {
				ant.Position++
				movementsMade = true
				continue
			}

			// Check if ant has reached the end
			if ant.Position >= len(ant.Path) {
				ant.Finished = true
				finished++
				roundMoves = append(roundMoves, fmt.Sprintf("L%d-%s", ant.ID, end))
				continue
			}

			// Check if next room is available
			if ant.Position < len(ant.Path) {
				nextRoom := ant.Path[ant.Position]
				if _, occupied := moves[nextRoom]; !occupied && nextRoom != end {
					ant.Position++
					moves[nextRoom] = ant.ID
					movementsMade = true
					roundMoves = append(roundMoves, fmt.Sprintf("L%d-%s", ant.ID, nextRoom))
				} else if nextRoom == end {
					// If next room is the end, multiple ants can occupy it
					ant.Position++
					ant.Finished = true
					finished++
					roundMoves = append(roundMoves, fmt.Sprintf("L%d-%s", ant.ID, end))
				}
			}
		}

		// Print all moves for this round
		if len(roundMoves) > 0 {
			fmt.Printf("%d:  %s\n", counter, strings.Join(roundMoves, " "))
		} else if movementsMade {
			fmt.Println()
		}

		if !movementsMade && finished < totalAnts {
			break
		}
	}
	return counter
}

func antDistribution(ants int, paths *[][]string) []int {
	pathLengths := make([]int, len(*paths))
	for i, p := range *paths {
		pathLengths[i] = len(p)
	}

	antsPerPath := make([]int, len(*paths))
	resPathLength := make([]int, len(*paths))
	copy(resPathLength, pathLengths)

	for ants > 0 {
		index := 0
		shortestPath := pathLengths[0]
		for i, path := range pathLengths {
			if path < shortestPath {
				shortestPath = path
				index = i
			}
		}
		antsPerPath[index]++
		pathLengths[index]++
		ants--
	}

	return antsPerPath
}

func rebuildGraph(farm *AntFarm, pathToRemove []string) *AntFarm {
	if len(pathToRemove) == 1 {
		return removeLink(farm, farm.Start, farm.End)
	}

	for i := 1; i < len(pathToRemove)-1; i++ {
		roomToRemove := pathToRemove[i]

		for roomName, room := range farm.Rooms {
			if roomName != roomToRemove {
				newLinks := []string{}
				for _, link := range room.Links {
					if link != roomToRemove {
						newLinks = append(newLinks, link)
					}
				}
				room.Links = newLinks
				farm.Rooms[roomName] = room
			}
		}

		delete(farm.Rooms, roomToRemove)
	}
	return farm
}

func copyGraph(farm *AntFarm) *AntFarm {
	newFarm := &AntFarm{
		Start: farm.Start,
		End:   farm.End,
		Ants:  farm.Ants,
		Rooms: make(map[string]*Room, len(farm.Rooms)),
	}

	// Copy all rooms and their links
	for name, room := range farm.Rooms {
		newLinks := make([]string, len(room.Links))
		copy(newLinks, room.Links)

		newFarm.Rooms[name] = &Room{
			Name:  room.Name,
			Links: newLinks,
		}
	}

	return newFarm
}

func removeLink(farm *AntFarm, fromNode, toNode string) *AntFarm {
	newFarm := copyGraph(farm)

	if room, exists := newFarm.Rooms[fromNode]; exists {
		newLinks := []string{}
		for _, link := range room.Links {
			if link != toNode {
				newLinks = append(newLinks, link)
			}
		}
		room.Links = newLinks
		newFarm.Rooms[fromNode] = room
	}

	if room, exists := newFarm.Rooms[toNode]; exists {
		newLinks := []string{}
		for _, link := range room.Links {
			if link != fromNode {
				newLinks = append(newLinks, link)
			}
		}
		room.Links = newLinks
		newFarm.Rooms[toNode] = room
	}
	return newFarm
}
