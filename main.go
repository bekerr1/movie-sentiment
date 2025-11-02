package main

import (
	"log"
	"sync"
)

type MovieSentimentEndpoint struct {
	movie    string
	endpoint string
}

//var movieReviewEndpoints = []MovieSentimentEndpoint{
//	{
//		movie:    "foo",
//		endpoint: "file:./test/generate/movie_foo_sentiment.txt",
//	},
//}

var movieReviewEndpoints = []MovieSentimentEndpoint{
	{
		movie:    "foo",
		endpoint: "http://nginx-file-server/movie_foo.txt",
	},
}

// Threads is number of concurrent threads to use for processing. NOTE: for testing its just 1
var Threads = 1

func main() {
	log.Println("Movie Sentiment Analysis Starting")

	var wg sync.WaitGroup
	var errors []error
	var errMutex sync.Mutex

	// Create Threads go routines to process movie sentiment endpoints
	// in parallel.
	tasks := make(chan MovieSentimentEndpoint, len(movieReviewEndpoints))
	for i := 0; i < Threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer func() {
				log.Printf("Goroutine %d stopping", id)
				wg.Done()
			}()
			log.Printf("Goroutine %d is running", id)
			for endpoint := range tasks {
				log.Printf("Goroutine %d processing endpoint: %s", id, endpoint)
				ms := NewMovieSentiment(endpoint.movie, endpoint.endpoint)
				if err := ms.Process(); err != nil {
					errMutex.Lock()
					errors = append(errors, err)
					errMutex.Unlock()
				}
			}
		}(i)
	}

	for _, endpoint := range movieReviewEndpoints {
		tasks <- endpoint
	}

	close(tasks)
	wg.Wait()

	if len(errors) > 0 {
		log.Fatalf("Encountered %v errors during processing", len(errors))
		for _, err := range errors {
			log.Printf("Error: %v", err)
		}
	}
}
