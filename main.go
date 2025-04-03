package main

import (
	"bufio"
	"fmt"
	"os"
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
		g = rebuildGraph(g, path[1])
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
	file, _ := os.Create("result.txt")
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Simulation loop, runs till all the ants have finished the course
	for finished < totalAnts {
		// Track movements for this turn
		moves := make(map[string]int) // room -> antID
		// Move each ant
		for _, ant := range ants {
			if ant.Finished {
				continue
			}
			// Check if ant should start moving
			if ant.Position < 0 {
				ant.Position++
				continue
			}
			// Check if ant has reached the end
			if ant.Position >= len(ant.Path) {
				ant.Finished = true
				finished++
				continue
			}
			// Get current room
			// currentRoom := ant.Path[ant.Position]
			// Check if next room is available
			if ant.Position < len(ant.Path) {
				nextRoom := ant.Path[ant.Position]
				if _, occupied := moves[nextRoom]; !occupied {
					ant.Position++
					moves[nextRoom] = ant.ID
					fmt.Printf("L%d-%s ", ant.ID, nextRoom)
					fmt.Fprintf(writer, "L%d-%s ", ant.ID, nextRoom)
				}
			} else {
				// Ant has reached the end
				ant.Finished = true
				finished++
				fmt.Printf("L%d-%s ", ant.ID, end)
				fmt.Fprintf(writer, "L%d-%s ", ant.ID, end)
			}
		}
		fmt.Println()
		fmt.Fprintf(writer, "\n")
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

func rebuildGraph(graph *AntFarm, nodeToRemove string) *AntFarm {
	// Delete references to this node from all other nodes
	for _, room := range graph.Rooms {
		newLinks := []string{}
		for _, link := range room.Links {
			if link != nodeToRemove {
				newLinks = append(newLinks, link)
			}
		}
		room.Links = newLinks
	}

	// Now delete the node itself
	delete(graph.Rooms, nodeToRemove)
	return graph
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
