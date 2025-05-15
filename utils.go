package main

import (
	"fmt"
	"runtime"
)

func printPerformance() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v KiB", m.Alloc/1024)
}
