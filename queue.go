package main

type Queue struct {
    elements []string
}

func NewQueue() *Queue {
    return &Queue{
        elements: make([]string, 0),
    }
}

func (q *Queue) Enqueue(element string) {
    q.elements = append(q.elements, element)
}

func (q *Queue) Dequeue() (string, bool) {
    if q.IsEmpty() {
        return "", false
    }
    
    element := q.elements[0]
    q.elements = q.elements[1:]
    return element, true
}

func (q *Queue) IsEmpty() bool {
    return len(q.elements) == 0
}
