package main

import (
	"fmt"
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
	ants := graph.Ants
	paths := [][]string{}

	path := bfs(graph.Start, graph.End, graph)

	// Check if the shortest path (path)
	if ants >= len(path) {
		// Use DFS
		path := dfs(graph.Start, graph.End, graph)
		paths = append(paths, path[1:])
		for len(path) > 0 {
			graph = rebuildGraph(graph, path)
			path = dfs(graph.Start, graph.End, graph)
			if len(path) != 0 {
				paths = append(paths, path[1:])
			}
		}
	} else {
		// Use BFS
		paths = append(paths, path[1:])
		for len(path) > 0 {
			graph = rebuildGraph(graph, path)
			path = bfs(graph.Start, graph.End, graph)
			if len(path) != 0 {
				paths = append(paths, path[1:])
			}
		}
	}

	fmt.Println(paths)
	antsPerPath := antDistribution(ants, &paths)
	movementSimulation(graph.End, ants, &antsPerPath, paths)
}

func movementSimulation(end string, totalAnts int, antsPerPath *[]int, paths [][]string) {
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

		// If no movements were made, but not all ants are finished,
		// we have a deadlock - break the loop
		if !movementsMade && finished < totalAnts {
			break
		}
	}
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

	// Special handling for start node: don't delete it but remove links to nodes in the path
	// if room, exists := newGraph.Rooms[pathToRemove[0]]; exists && pathToRemove[0] == graph.Start {
	// 	newLinks := []string{}
	// 	for _, link := range room.Links {
	// 		// Keep links that don't point to nodes in the path
	// 		if link != pathToRemove[1] {
	// 			newLinks = append(newLinks, link)
	// 		}
	// 	}
	// 	room.Links = newLinks
	// 	newGraph.Rooms[pathToRemove[0]] = room
	// }

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

func bfs(start, end string, graph *AntFarm) []string {
	queue := NewQueue()
	queue.Enqueue(start)

	visited := make(map[string]bool)
	visited[start] = true

	path := make(map[string]string)

	for !queue.IsEmpty() {
		node, _ := queue.Dequeue()

		if node == end {
			tPath := []string{end}
			current := end
			for path[current] != "" {
				current = path[current]
				tPath = append([]string{current}, tPath...)
			}
			return tPath
		}

		neighbors := graph.Rooms[node].Links
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue.Enqueue(neighbor)
				path[neighbor] = node
			}
		}
	}

	return []string{}
}

func dfs(start, end string, graph *AntFarm) []string {
	stack := &Stack{}
	stack.Push(start)

	visited := make(map[string]bool)
	visited[start] = true
	path := make(map[string]string)

	for stack.Size() > 0 {
		node, _ := stack.Pop()

		if node == end {
			tPath := []string{end}
			current := end
			for path[current] != "" {
				current = path[current]
				tPath = append([]string{current}, tPath...)
			}
			return tPath
		}

		neighbors := graph.Rooms[node].Links

		for i := len(neighbors) - 1; i >= 0; i-- {
			neighbor := neighbors[i]
			if !visited[neighbor] {
				stack.Push(neighbor)
				path[neighbor] = node
				visited[neighbor] = true
			}
		}
	}
	return []string{}
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
