package main

import (
	"fmt"
)

const (
	maxWorkers = 5
	numJobs    = 10
)

func main() {
	pool := NewPool(maxWorkers)
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for j := range jobs {
		j := j
		pool.Submit(func() {
			results <- j * 2
		})
	}

	for a := 1; a <= numJobs; a++ {
		fmt.Println(<-results)
	}
	close(results)
}
