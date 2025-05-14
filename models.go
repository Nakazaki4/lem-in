package main

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

type Room struct {
	Name  string
	Links []string
	Used  bool
}

type AntFarm struct {
	Ants  int
	Rooms map[string]*Room
	Start string
	End   string
}

type PathNode struct {
	Name   string
	Parent *PathNode
}

