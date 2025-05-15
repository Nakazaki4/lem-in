package main

func getAllPossiblePathsBfs(farm *AntFarm) [][]string {
	var paths [][]string
	queue := Queue{}
	queue.Enqueue([]string{farm.Start})

	for !queue.IsEmpty() {
		path := queue.Dequeue()
		current := path[len(path)-1]
		if current == farm.End {
			paths = append(paths, path[1:])
			continue
		}

		visited := make(map[string]bool)

		for _, room := range path {
			visited[room] = true
		}

		currentRoom, exists := farm.Rooms[current]
		if !exists || currentRoom == nil {
			// Skip this path if the room doesn't exist
			continue
		}

		for _, neighbor := range farm.Rooms[current].Links {
			if visited[neighbor] {
				continue
			}

			visited[neighbor] = true

			newPath := make([]string, len(path))
			copy(newPath, path)
			newPath = append(newPath, neighbor)

			// -- Use this for big graphs --
			// hasOverLap := false
			// newRooms := make(map[string]bool)
			// for _, room := range newPath[1:] {
			// 	newRooms[room] = true
			// }

			// for _, existingPath := range queue.elements {
			// 	for _, room := range existingPath {
			// 		if newRooms[room] {
			// 			hasOverLap = true
			// 			break
			// 		}
			// 	}
			// 	if hasOverLap {
			// 		break
			// 	}
			// }

			// if !hasOverLap {
			// 	queue.Enqueue(newPath)
			// }

			// -- Use this for small graphs --
			queue.Enqueue(newPath)
		}
	}
	return paths
}

func findShortestPath(farm *AntFarm, startRoom string, endRoom string) []string {
	if startRoom == endRoom {
		return []string{startRoom}
	}

	queue := []string{startRoom}

	visited := make(map[string]bool)
	visited[startRoom] = true

	parent := make(map[string]string)

	// BFS traversal
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == endRoom {
			path := []string{current}
			for current != startRoom {
				current = parent[current]
				path = append([]string{current}, path...)
			}
			return path
		}

		if room, ok := farm.Rooms[current]; ok {
			if room.Name != farm.Start {
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

	return nil
}