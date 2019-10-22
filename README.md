# Golang Code Sample

## Overview and background

This is a short program that generates random prime numbers within a given range and through a specified number of workers. I developed this code sample from when I ran a workshop teaching university students about Golang. The topic was on experimenting with the fan-out and fan-in pattern of workers and channels using Golang's concurrency model.

The focus was not prime number generation, it was just selected as a small problem domain to work with. Extension of this code can involve handling multiple layers of complex concurrency (e.g processing batches of jobs within a system by using wait groups, passing channels within channels, worker pools, etc.)

## Usage

Takes in three arguments:
- p = Number of prime numbers to generate
- r = Range of random numbers to be used as an input stream, values from 0 to r
- n = Number of workers to be used to process the input  

Example usage:
`go run main.go p=15 r=10000000 n=10`

## Code details

Process followed to generate prime numbers:
1. Generate a generic input stream (channel) which gets random values by calling a getter that is passed in as a param
2. Convert input stream values to a stream of the desired type (int in our case)
3. Create a worker that gets prime numbers from an input. Fan out the number of worker channels to be used during processing
4. Combine/multiplex result of all worker channels. Values fanned in onto a single stream
5. Result is a single stream containing prime number outputs from all workers.

- Interfaces are used in a few places to make the code extensible (for purposes other than prime number generation)
- Code should be split up into seperate files when extending support for different input stream types and different types of workers (other than integers and prime number generation).  

```