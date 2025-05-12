package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

type Ant struct {
	ID       int
	PathID   int
	Position int
	Path     []string
	Finished bool
}

type BestSolution struct {
	Turns        int
	Steps        int
	Group        [][]string
	Distribution []int
}

func printPerformance() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v KiB", m.Alloc/1024)
	fmt.Printf("\tTotalAlloc = %v KiB", m.TotalAlloc/1024)
	fmt.Printf("\tSys = %v KiB\n", m.Sys/1024)
}

func main() {
	fileName := os.Args[1]
	file, _ := os.Open(fileName)
	graph := ParseFile(file)
	newApproach(graph)
	printPerformance()
}

func newApproach(graph *AntFarm) {
	// Get all possible paths first
	allPaths := make([][]string, 0)
	for _, neighbor := range graph.Rooms[graph.Start].Links {
		if neighbor != graph.End {
			allPaths = append(allPaths, findShortestPath(graph, neighbor, graph.End))
		}
	}
	// allPaths := getAllPossiblePathsBfs(graph)
	// disjointPaths := findDisjointPaths(allPaths)
	// fmt.Println(disjointPaths)
	// This will hold all combinations of compatible paths
	allCombinations := [][][]string{}

	// Try each path as a starting point
	for _, path := range allPaths {
		if len(path) != 0 && len(path) <= graph.Ants {

			// Start a new combination with this path
			combo := [][]string{path}
			allCombinations = append(allCombinations, combo)

			newGraph := rebuildGraph(copyGraph(graph), path)
			// Find all compatible paths recursively
			findCompatiblePathsRecursive(newGraph, &combo, &allPaths, &allCombinations)
		}
	}
	// // Map to store evaluation metrics for each combination
	stepTurns := make(map[int][]int)

	for i, combin := range allCombinations {
		antsPerPath := antDistribution(graph.Ants, &combin)

		// Calculate the turns using the first path's length and ants
		firstPathLength := len(combin[0])
		firstPathAnts := antsPerPath[0]
		// this calculates turns
		tempT := firstPathLength - 1 + firstPathAnts

		totalSteps := 0
		// this calculates steps
		for j, path := range combin {
			totalSteps += antsPerPath[j] * len(path)
		}

		stepTurns[i] = []int{totalSteps, tempT}
	}

	// Find the best combination using MinSteps
	bestIndex := MinSteps(stepTurns)
	bestCombo := allCombinations[bestIndex]
	bestDistribution := antDistribution(graph.Ants, &bestCombo)

	movementSimulation(graph.End, graph.Ants, &bestDistribution, bestCombo)
}

func findShortestPath(graph *AntFarm, startNode string, endNode string) []string {
	// Use the graph's end if not specified
	if endNode == "" {
		endNode = graph.End
	}

	// Check if start and end are the same
	if startNode == endNode {
		return []string{startNode}
	}

	// Queue for BFS
	queue := []string{startNode}

	// Keep track of visited nodes to avoid cycles
	visited := make(map[string]bool)
	visited[startNode] = true

	// Store parent of each node to reconstruct the path
	parent := make(map[string]string)

	// BFS traversal
	for len(queue) > 0 {
		// Dequeue the first node
		current := queue[0]
		queue = queue[1:]

		// Check if we've reached the end
		if current == endNode {
			// Reconstruct the path
			path := []string{current}
			for current != startNode {
				current = parent[current]
				path = append([]string{current}, path...)
			}
			return path
		}

		// Explore neighbors
		if room, ok := graph.Rooms[current]; ok {
			if room.Name != graph.Start {
				for _, neighbor := range room.Links {
					if !visited[neighbor] {
						visited[neighbor] = true
						parent[neighbor] = current
						queue = append(queue, neighbor)
					}
				}
			}
		}
	}

	// No path found
	return nil
}

func findDisjointPaths(allPaths [][]string) [][]string {
	// Sort paths by length (shortest first)
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i]) < len(allPaths[j])
	})

	selected := make([][]string, 0)
	usedRooms := make(map[string]bool)

	for _, path := range allPaths {
		conflict := false
		// Check intermediate nodes (exclude start and end)
		for _, room := range path[1 : len(path)-1] {
			if usedRooms[room] {
				conflict = true
				break
			}
		}

		if !conflict {
			// Mark nodes as used
			for _, node := range path[1 : len(path)-1] {
				usedRooms[node] = true
			}
			selected = append(selected, path)
		}
	}

	return selected
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

func findCompatiblePathsRecursive(modifiedGraph *AntFarm, currentGroup *[][]string, allPaths *[][]string, allCombinations *[][][]string) {
	possiblePaths := getAllPossiblePathsBfs(modifiedGraph)
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
		if !isCompatibleWithComb(currentGroup, &p) {
			continue
		}
		if len(newGroup) > modifiedGraph.Ants {
			continue
		}

		if len(newGroup) > 0 {
			*allCombinations = append(*allCombinations, newGroup)
		}

		// Further modify the graph and continue recursively
		newGraph := rebuildGraph(copyGraph(modifiedGraph), path)
		findCompatiblePathsRecursive(newGraph, &newGroup, allPaths, allCombinations)
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

func getAllPossiblePathsBfs(graph *AntFarm) [][]string {
	var paths [][]string
	queue := Queue{}
	queue.Enqueue([]string{graph.Start})

	for !queue.IsEmpty() {
		path := queue.Dequeue()
		current := path[len(path)-1]
		if current == graph.End {
			paths = append(paths, path[1:])
			continue
		}

		visited := make(map[string]bool)

		for _, room := range path {
			visited[room] = true
		}

		currentRoom, exists := graph.Rooms[current]
		if !exists || currentRoom == nil {
			// Skip this path if the room doesn't exist
			continue
		}

		for _, neighbor := range graph.Rooms[current].Links {
			if visited[neighbor] {
				continue
			}

			visited[neighbor] = true

			newPath := make([]string, len(path))
			copy(newPath, path)
			newPath = append(newPath, neighbor)

			queue.Enqueue(newPath)
		}
	}
	return paths
}

// func isJoined(newPath [][]string) bool {
// }

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

func rebuildGraph(graph *AntFarm, pathToRemove []string) *AntFarm {
	if len(pathToRemove) == 1 {
		fmt.Println("REMOVE PATH")
		return removeLink(graph, graph.Start, graph.End)
	}

	for i := 1; i < len(pathToRemove)-1; i++ {
		nodeToRemove := pathToRemove[i]

		for roomName, room := range graph.Rooms {
			if roomName != nodeToRemove {
				newLinks := []string{}
				for _, link := range room.Links {
					if link != nodeToRemove {
						newLinks = append(newLinks, link)
					}
				}
				room.Links = newLinks
				graph.Rooms[roomName] = room
			}
		}

		delete(graph.Rooms, nodeToRemove)
	}
	return graph
}

func copyGraph(graph *AntFarm) *AntFarm {
	newGraph := &AntFarm{
		Start: graph.Start,
		End:   graph.End,
		Ants:  graph.Ants,
		Rooms: make(map[string]*Room, len(graph.Rooms)),
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
