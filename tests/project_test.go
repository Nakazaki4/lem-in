package lemin

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestLemin(t *testing.T) {
	tests := []struct {
		Name      string
		Turns     int
		TimeLimit time.Duration
	}{
		{
			Name:  "example00.txt",
			Turns: 6,
		},
		{
			Name:  "example01.txt",
			Turns: 8,
		},
		{
			Name:  "example02.txt",
			Turns: 11,
		},
		{
			Name:  "example03.txt",
			Turns: 6,
		},
		{
			Name:  "example04.txt",
			Turns: 6,
		},
		{
			Name:  "example05.txt",
			Turns: 8,
		},
		{
			Name:      "example06.txt",
			TimeLimit: 90 * time.Second,
		},
		{
			Name:      "example07.txt",
			TimeLimit: 150 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".", tt.Name)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Timewise
			if tt == tests[len(tests)-1] || tt == tests[len(tests)-2] {
				start := time.Now()
				err := cmd.Run()
				finish := time.Now()
				if err == nil {
					t.Errorf("Failed to run %s", tt.Name)
					return
				}
				elapsed := finish.Second() - start.Second()
				if elapsed > int(tt.TimeLimit) {
					t.Errorf("took too much time %d", elapsed)
				}
				return
			}

			err := cmd.Run()

			if err == nil {
				t.Errorf("Failed to run %s", tt.Name)
				return
			}

			output := stdout.String()

			turns := countTurns(output)

			if turns > uint(tt.Turns) {
				t.Errorf("Too many turns on the file %s", tt.Name)
			}
		})
	}
}

func countTurns(result string) uint {
	splitted := strings.Split(result, "\n")
	return uint(len(splitted))
}
