package main

import (
	"fmt"
	"os"
	"slices"
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

	
	for allPaths >  {

	}
	// This will hold all combinations of compatible paths
	allCombinations := [][][]string{}

	// Try each path as a starting point
	for _, path := range allPaths {
		// Start a new combination with this path
		combo := [][]string{path}
		allCombinations = append(allCombinations, combo)

		// Create a graph with this path removed
		modifiedGraph := rebuildGraph(copyGraph(graph), path)

		// Find all compatible paths recursively
		findCompatiblePathsRecursive(modifiedGraph, combo, allPaths, &allCombinations)
	}

	for i, combin := range allCombinations {
		fmt.Println(i, "-->", combin)
	}

	// Map to store evaluation metrics for each combination
	stepTurns := make(map[int][]int)

	// Now evaluate each combination using BestGroup's approach
	for i, combin := range allCombinations {
		// Calculate ant distribution for this combination
		antsPerPath := antDistribution(graph.Ants, &combin)

		// Calculate the BestGroup metrics
		// Calculate the turns using the first path's length and ants
		firstPathLength := len(combin[0])
		firstPathAnts := antsPerPath[0]
		// this calculates turns
		tempT := firstPathLength - 1 + firstPathAnts

		totalSteps := 0 // Total weighted path length
		// this calculates steps
		for j, path := range combin {
			totalSteps += antsPerPath[j] * len(path)
		}

		// Store the metrics for this combination
		stepTurns[i] = []int{totalSteps, tempT}
	}

	// Find the best combination using MinSteps
	bestIndex := MinSteps(stepTurns)
	bestCombo := allCombinations[bestIndex]
	bestDistribution := antDistribution(graph.Ants, &bestCombo)

	// Get metrics for the best combination
	bestMetrics := stepTurns[bestIndex]

	fmt.Printf("Best solution takes %d turns\n", bestMetrics[1])
	fmt.Printf("Total weighted path length: %d\n", bestMetrics[0])
	fmt.Printf("Using paths: %v\n", bestCombo)
	fmt.Printf("With ant distribution: %v\n", bestDistribution)
}

// Implementation of MinSteps function
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

func findCompatiblePathsRecursive(modifiedGraph *AntFarm, currentCombo [][]string, allPaths [][]string, allCombinations *[][][]string) {
	compatiblePaths := getAllPossiblePaths(modifiedGraph)

	for _, path := range compatiblePaths {
		newCombo := make([][]string, len(currentCombo))
		copy(newCombo, currentCombo)
		// we should check if the path we're about to append doesn't intersect with any paths in the previous combination
		path = path[:len(path)-1]
		if !isCompatibleWithComb(&currentCombo, &path) {
			continue
		}
		newCombo = append(newCombo, path)

		// Add this new combination
		*allCombinations = append(*allCombinations, newCombo)

		// Further modify the graph and continue recursively
		newGraph := rebuildGraph(copyGraph(modifiedGraph), path)
		findCompatiblePathsRecursive(newGraph, newCombo, allPaths, allCombinations)
	}
}

func isCompatibleWithComb(combination *[][]string, pathToAppend *[]string) bool {
	for _, path := range *combination {
		for i := range path {
			if slices.Contains(*pathToAppend, path[i]) {
				return false
			}
		}
	}
	return true
}

func getAllPossiblePaths(graph *AntFarm) [][]string {
	var paths [][]string
	var dfs func(path []string, visited map[string]bool)

	dfs = func(path []string, visited map[string]bool) {
		current := path[len(path)-1]
		if current == graph.End {
			paths = append(paths, path[1:])
			return
		}

		for _, neighbor := range graph.Rooms[current].Links {
			if !visited[neighbor] {
				visited[neighbor] = true
				dfs(append(path, neighbor), visited)
				visited[neighbor] = false // Backtrack
			}
		}
	}

	visited := make(map[string]bool)
	visited[graph.Start] = true
	dfs([]string{graph.Start}, visited)
	return paths
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
	if len(pathToRemove) == 1 {
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
