package main

import (
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

const (
	DEFAULT_NUM_PRIMES  = 10
	DEFAULT_NUM_RANGE   = 100000
	DEFAULT_NUM_WORKERS = 8
)

// An experimental program that:
// - Finds P prime numbers
// - From a stream of random input values, within range 0 to R
// - Using N workers that operate on the stream
// Usage: go run main.go -p=10 -r=1000000 -n=8
func main() {
	numPrimes := flag.Int("p", DEFAULT_NUM_PRIMES, "Number of prime numbers to generate")
	numRange := flag.Int64("r", DEFAULT_NUM_RANGE, "Range of numbers to search from")
	numWorkers := flag.Int("n", DEFAULT_NUM_WORKERS, "Number of workers to concurrently process values")
	flag.Parse()
	fmt.Printf("Generating %d random prime numbers within range 0-%d...\n", *numPrimes, *numRange)
	fmt.Printf("Creating %d workers...\n", *numWorkers)

	done := make(chan interface{})
	defer close(done)
	start := time.Now()

	// Generate an input stream of random ints
	valueStream := createValueStream(done, randVal(*numRange))
	intStream := valuesToIntStream(done, valueStream)

	// Set workers that get prime numbers from input. Fan out the workers
	workers := make([]<-chan interface{}, *numWorkers)
	for i := 0; i < *numWorkers; i++ {
		workers[i] = primeNumberWorker(done, intStream)
	}

	// Multiplex result from all workers, fanning in the results to a single stream of prime numbers
	primeNumberFinder := reduceWorkers(done, workers...)
	primeNumberStream := createResultStream(done, primeNumberFinder, *numPrimes)

	fmt.Println("Prime numbers generated:")
	for num := range primeNumberStream {
		fmt.Printf("%d\n", num)
	}

	fmt.Printf("Duration: %v\n", time.Since(start))
}

// createResultStream gets a stream containing the number of specified items from a given input stream (number of prime numbers to generate in our usage)
func createResultStream(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
	result := make(chan interface{})
	go func() {
		defer close(result)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case result <- <-valueStream:
			}
		}
	}()
	return result
}

// reduceWorkers takes a set of generic channels (worker channels containing prime numbers in our usage) and multiplexes their streams into a single stream
func reduceWorkers(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	reducedStream := make(chan interface{})

	// Forwards output of given channel to one stream
	reduceChan := func(workerChannel <-chan interface{}) {
		defer wg.Done()
		for item := range workerChannel {
			select {
			case <-done:
				return
			case reducedStream <- item:
			}
		}
	}

	// Combine output of all worker channels. Wait until all items are processed
	wg.Add(len(channels))
	for _, wc := range channels {
		go reduceChan(wc)
	}
	go func() {
		wg.Wait()
		close(reducedStream)
	}()

	return reducedStream
}

// primeNumberWorker reads an input stream of numbers and outputs a stream of prime numbers it finds
func primeNumberWorker(done <-chan interface{}, intStream <-chan int64) <-chan interface{} {
	primeNumStream := make(chan interface{})
	go func() {
		defer close(primeNumStream)
		for num := range intStream {
			// Check if prime number found
			if big.NewInt(num).ProbablyPrime(0) {
				select {
				case <-done:
					return
				case primeNumStream <- num:
				}
			}
		}
	}()
	return primeNumStream
}

// createValueStream gets values from a specified getter, and queues the result on a stream (generic result type)
func createValueStream(done <-chan interface{}, getValue func() interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case valStream <- getValue(): // Call getter, place result on stream
			}
		}
	}()
	return valStream
}

// valuesToIntStream returns a channel, which converts a generic stream to an explicit type (type int in our case)
func valuesToIntStream(done <-chan interface{}, vals <-chan interface{}) <-chan int64 {
	intStream := make(chan int64)
	go func() {
		defer close(intStream)
		for item := range vals {
			select {
			case <-done:
				return
			case intStream <- item.(int64):
			}
		}
	}()
	return intStream
}

// randVal returns a function, which returns a generic value (a random int in our case)
func randVal(num int64) func() interface{} {
	return func() interface{} {
		return rand.Int63n(num)
	}
}
