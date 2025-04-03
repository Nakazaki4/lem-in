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
	Path     []string // The path that this ant is following
	Finished bool
}

func main() {
	fileName := os.Args[1]
	f, _ := os.Open(fileName)
	g := ParseFile(f)
	ants := g.Ants
	paths := [][]string{}
	path := BFS(g.Start, g.End, g)
	paths = append(paths, path[1:])
	for len(path) > 0 {
		g = rebuildGraph(g, path)
		path = BFS(g.Start, g.End, g)
		if len(path) != 0 {
			paths = append(paths, path[1:])
		}
	}
	antsPerPath := antDistribution(ants, &paths)
	movementSimulation(g.End, ants, &antsPerPath, paths)
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

	// Simulation loop, runs till all the ants have finished the course
	for finished < totalAnts {
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
			fmt.Println(strings.Join(roundMoves, " "))
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
	newGraph := copyGraph(graph)

	// Remove all internal nodes in the path, exclude start and end nodes
	for i := 1; i < len(pathToRemove)-1; i++ {
		nodeToRemove := pathToRemove[i]

		// First, remove all links to this node from other rooms
		for roomName, room := range newGraph.Rooms {
			if roomName != nodeToRemove {
				// Filter out links to the node being removed
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

		// Then delete the node itself
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

func BFS(start, end string, graph *AntFarm) []string {
	queue := NewQueue()
	queue.Enqueue(start)

	visited := make(map[string]bool)
	visited[start] = true
	path := make(map[string]string)

	for !queue.IsEmpty() {
		node, _ := queue.Dequeue()
		neighbors := graph.Rooms[node].Links

		if node == end {
			tPath := []string{end}
			current := end
			for path[current] != "" {
				current = path[current]
				tPath = append([]string{current}, tPath...)
			}
			return tPath
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				queue.Enqueue(neighbor)
				path[neighbor] = node
				visited[neighbor] = true
			}
		}
	}
	return []string{}
}
