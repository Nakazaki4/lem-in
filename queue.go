package main

import "fmt"

// Queue is a simple FIFO queue implementation
type Queue struct {
	items []string
}

// NewQueue creates a new empty queue
func NewQueue() *Queue {
	return &Queue{
		items: []string{},
	}
}

// Enqueue adds an item to the end of the queue
func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the item at the front of the queue
func (q *Queue) Dequeue() (string, error) {
	if len(q.items) == 0 {
		return "", fmt.Errorf("queue is empty")
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Front returns the item at the front of the queue without removing it
func (q *Queue) Front() (string, error) {
	if len(q.items) == 0 {
		return "", fmt.Errorf("queue is empty")
	}
	return q.items[0], nil
}

// IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

// Size returns the number of items in the queue
func (q *Queue) Size() int {
	return len(q.items)
}