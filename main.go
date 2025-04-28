package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type Ant struct {
	ID       int
	PathID   int
	Position int
	Path     []string
	Finished bool
}

func main() {
	fileName := os.Args[1]
	file, _ := os.Open(fileName)
	graph := ParseFile(file)
	newApproach(graph)
}

func newApproach(graph *AntFarm) {
	// Get all possible paths first
	allPaths := getAllPossiblePaths(graph)
	// This will hold all combinations of compatible paths
	allCombinations := [][][]string{}

	// Try each path as a starting point
	for _, path := range allPaths {
		// Start a new combination with this path
		combo := [][]string{path[1:]}
		allCombinations = append(allCombinations, combo)

		// Create a graph with this path removed
		modifiedGraph := rebuildGraph(copyGraph(graph), path[1:])

		// Find all compatible paths recursively
		findCompatiblePathsRecursive(modifiedGraph, combo, allPaths, &allCombinations)
	}

	// Now that we have all possible combinations we have to try to send ants through all possible combinations to get the minimu steps a path will give
	// This is ditribution
	targetSteps := math.MaxInt16
	var bestCombo [][]string
	var bestDistribution []int
	shortestPathLength := math.MaxInt64
	for _, combin := range allCombinations {
		// Calculate ant distribution for this combination
		antsPerPath := antDistribution(graph.Ants, &combin)

		// Simulate movement and get steps needed
		steps := movementSimulation(graph.End, graph.Ants, &antsPerPath, combin)

		// Calculate total path length (if this matters for tiebreaking)
		totalPathLength := 0
		for _, path := range combin {
			totalPathLength += len(path)
		}

		// First priority: minimize steps
		if steps < targetSteps {
			targetSteps = steps
			bestCombo = combin
			bestDistribution = antsPerPath
			shortestPathLength = totalPathLength
		} else if steps == targetSteps && totalPathLength < shortestPathLength {
			bestCombo = combin
			bestDistribution = antsPerPath
			shortestPathLength = totalPathLength
		}
	}

	fmt.Printf("Best solution takes %d steps\n", targetSteps)
	fmt.Printf("Total path length: %d\n", shortestPathLength)
	fmt.Printf("Using paths: %v\n", bestCombo)
	fmt.Printf("With ant distribution: %v\n", bestDistribution)
}

func findCompatiblePathsRecursive(modifiedGraph *AntFarm, currentCombo [][]string, allPaths [][]string, allCombinations *[][][]string) {
	// Find all possible paths in this modified graph
	compatiblePaths := getAllPossiblePaths(modifiedGraph)

	// For each compatible path
	for _, path := range compatiblePaths {
		// Add this path to our combination
		newCombo := make([][]string, len(currentCombo))
		copy(newCombo, currentCombo)
		newCombo = append(newCombo, path[1:])

		// Add this new combination
		*allCombinations = append(*allCombinations, newCombo)

		// Further modify the graph and continue recursively
		newGraph := rebuildGraph(copyGraph(modifiedGraph), path[1:])
		findCompatiblePathsRecursive(newGraph, newCombo, allPaths, allCombinations)
	}
}

func getAllPossiblePaths(graph *AntFarm) [][]string {
	// Store all discovered paths
	allPaths := [][]string{}

	// Track visited nodes to avoid cycles
	visited := make(map[string]bool)

	// Current path being explored
	currentPath := []string{}

	// Recursive DFS helper function
	var dfsHelper func(current string, end string)

	dfsHelper = func(current string, end string) {
		// Mark current node as visited
		visited[current] = true

		// Add current node to path
		currentPath = append(currentPath, current)

		// If we reached the end, we found a path
		if current == end {
			// Create a copy of the current path to avoid reference issues
			pathCopy := make([]string, len(currentPath))
			copy(pathCopy, currentPath)
			allPaths = append(allPaths, pathCopy)
		} else {
			// Try all neighbors
			for _, neighbor := range graph.Rooms[current].Links {
				if !visited[neighbor] {
					dfsHelper(neighbor, end)
				}
			}
		}

		// Backtrack - remove current node from path and mark it as unvisited
		visited[current] = false
		currentPath = currentPath[:len(currentPath)-1]
	}

	// Start DFS from the start node
	dfsHelper(graph.Start, graph.End)

	return allPaths
}

func movementSimulation(end string, totalAnts int, antsPerPath *[]int, paths [][]string) int {
	ants := make([]*Ant, totalAnts)
	antID := 1

	// create ants and assign them to paths (this brings how many ants are in each path)
	for pathID, count := range *antsPerPath {
		// This takes the number of ants in a path and gives each ant a ID and a ID of the path it belomngs too
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

	// Simulation loop, runs till all the ants have finished the course
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
			fmt.Printf("NUMBER: %d %s\n", counter, strings.Join(roundMoves, " "))
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
	// Length of each path
	pathLengths := make([]int, len(*paths))

	for i, p := range *paths {
		pathLengths[i] = len(p)
	}
	resPathLength := make([]int, len(*paths))
	copy(resPathLength, pathLengths)

	for ants > 0 {
		shortestPath := pathLengths[0]
		index := 0
		// To find the shortest path
		for i, path := range pathLengths {
			if path < shortestPath {
				shortestPath = path
				index = i
			}
		}
		// Assign an ant to the shortest path
		pathLengths[index]++
		ants--
	}

	// Shows how many ants have been assigned to each path
	antsPerPath := make([]int, len(*paths))
	for i, p := range pathLengths {
		antsPerPath[i] = p - resPathLength[i]
	}
	return antsPerPath
}

func rebuildGraph(graph *AntFarm, pathToRemove []string) *AntFarm {
	if len(pathToRemove) == 2 {
		return removeLink(graph, graph.Start, graph.End)
	}

	newGraph := copyGraph(graph)

	for i := 1; i < len(pathToRemove)-1; i++ {
		nodeToRemove := pathToRemove[i]

		for roomName, room := range newGraph.Rooms {
			if roomName != nodeToRemove {
				newLinks := []string{}
				for _, link := range room.Links {
					if link != nodeToRemove {
						newLinks = append(newLinks, link)
					}
				}
				room.Links = newLinks
				newGraph.Rooms[roomName] = room
			}
		}

		delete(newGraph.Rooms, nodeToRemove)
	}
	return newGraph
}

func copyGraph(graph *AntFarm) *AntFarm {
	newGraph := &AntFarm{
		Start: graph.Start,
		End:   graph.End,
		Ants:  graph.Ants,
		Rooms: make(map[string]*Room),
	}

	// Copy all rooms and their links
	for name, room := range graph.Rooms {
		newLinks := make([]string, len(room.Links))
		copy(newLinks, room.Links)

		newGraph.Rooms[name] = &Room{
			Name:  room.Name,
			Links: newLinks,
		}
	}

	return newGraph
}

func removeLink(graph *AntFarm, fromNode, toNode string) *AntFarm {
	newGraph := copyGraph(graph)

	if room, exists := newGraph.Rooms[fromNode]; exists {
		newLinks := []string{}
		for _, link := range room.Links {
			if link != toNode {
				newLinks = append(newLinks, link)
			}
		}
		room.Links = newLinks
		newGraph.Rooms[fromNode] = room
	}

	if room, exists := newGraph.Rooms[toNode]; exists {
		newLinks := []string{}
		for _, link := range room.Links {
			if link != fromNode {
				newLinks = append(newLinks, link)
			}
		}
		room.Links = newLinks
		newGraph.Rooms[toNode] = room
	}
	return newGraph
}
