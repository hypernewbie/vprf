package main

import (
	"os"
	"strconv"
	"time"
)

func innerLoop() int {
	total := 0
	for i := 0; i < 3_000_000; i++ {
		total += i * i
	}
	return total
}

func outer() []int {
	results := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		results = append(results, innerLoop())
	}
	return results
}

func main() {
	duration := 4.0
	if len(os.Args) > 1 {
		if parsed, err := strconv.ParseFloat(os.Args[1], 64); err == nil {
			duration = parsed
		}
	}
	end := time.Now().Add(time.Duration(duration * float64(time.Second)))
	for time.Now().Before(end) {
		_ = outer()
	}
}
