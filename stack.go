package main

import "fmt"

type Stack struct {
	items []string
	name string
}

func (s *Stack) NewStack(name string) *Stack{
	return &Stack{
		items: []string{},
		name: name,
	}
}

// Adds the item to the top of the stack
func (s *Stack) Push(item string){
	s.items = append([]string{item}, s.items...)
}

// Removes the item in the top
func (s *Stack) Pop() (string, error) {
	if len(s.items)==0{
		return "", fmt.Errorf("stack is empty")
	}
	v := s.items[0]
	s.items = s.items[1:]
	return v, nil
}

// Returns the item in the top without deleting it
func (s *Stack) Peek() (string, error){
	if len(s.items) == 0{
		return "", fmt.Errorf("stack is empty")
	}
	return s.items[0], nil
}

func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *Stack) Size() int {
	return len(s.items)
}

