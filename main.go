package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func calculateEvenSum(numbers []int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	sum := 0
	for _, num := range numbers {
		if num%2 == 0 {
			sum += num
		}
	}
	results <- sum
}

func generateLargeIntSlice(size int) []int {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	numbers := make([]int, size)
	for i := 0; i < size; i++ {
		numbers[i] = r.Intn(1000)
	}
	return numbers
}

func calculateEvenSumSequential(numbers []int) (int, time.Duration) {
	start := time.Now()
	sum := 0
	for _, num := range numbers {
		if num%2 == 0 {
			sum += num
		}
	}
	return sum, time.Since(start)
}

func main() {
	const sliceSize = 1000000
	const numWorkers = 4

	fmt.Printf("Generating slice of %d integers...\n", sliceSize)
	numbers := generateLargeIntSlice(sliceSize)

	fmt.Println("Calculating sum sequentially...")
	sequentialSum, sequentialTime := calculateEvenSumSequential(numbers)

	fmt.Printf("Calculating sum concurrently using %d workers...\n", numWorkers)
	concurrentSum, concurrentTime := calculateEvenSumConcurrent(numbers, numWorkers)

	fmt.Printf("\n=== RESULTS ===\n")
	fmt.Printf("Slice size: %d integers\n", sliceSize)
	fmt.Printf("Number of workers: %d\n", numWorkers)
	fmt.Printf("Sequential sum: %d (took %v)\n", sequentialSum, sequentialTime.Seconds())
	fmt.Printf("Concurrent sum: %d (took %v)\n", concurrentSum, concurrentTime.Seconds())
	fmt.Printf("Speedup: %.2fx faster\n", float64(sequentialTime)/float64(concurrentTime))

	if sequentialSum == concurrentSum {
		fmt.Println("✅ Results match!")
	} else {
		fmt.Println("❌ Results don't match!")
	}
}

func calculateEvenSumConcurrent(numbers []int, numWorkers int) (int, time.Duration) {
	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan int, numWorkers)

	chunkSize := len(numbers) / numWorkers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		start := i * chunkSize
		end := start + chunkSize

		if i == numWorkers-1 {
			end = len(numbers)
		}

		go calculateEvenSum(numbers[start:end], results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	totalSum := 0
	for partialSum := range results {
		totalSum += partialSum
	}

	return totalSum, time.Since(start)
}
