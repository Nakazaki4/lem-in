package main

type Queue struct {
	elements [][]string
}

func (q *Queue) Enqueue(path []string) {
	q.elements = append(q.elements, path)
}

func (q *Queue) Dequeue() []string {
	if q.IsEmpty() {
		return []string{}
	}
	path := q.elements[0]
	q.elements = q.elements[1:]
	return path
}

func (q *Queue) Peek() []string {
	if q.IsEmpty() {
		return []string{}
	}
	return q.elements[0]
}

func (q *Queue) IsEmpty() bool {
	return len(q.elements) == 0
}

func (q *Queue) Size() int {
	return len(q.elements)
}
