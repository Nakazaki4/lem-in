package main

type Ant struct {
	ID       int
	PathID   int
	Position int
	Path     []string
	Finished bool
}

type Room struct {
	Name  string
	Links []string
}

type AntFarm struct {
	Ants  int
	Rooms map[string]*Room
	Start string
	End   string
}
