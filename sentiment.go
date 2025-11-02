package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
)

type MovieSentiment struct {
	// Movie identifier for metrics/logging
	movie string
	// URL we started processing from
	seedURL string
	// Currently gathered negative sentiment for this movie
	negativeSentiment      []string
	negativeSentimentBytes int
	// max size negative sentiment can grow before we flush
	maxNegativeSentimentSizeBytes int
	// File to write negative sentiment reviews to. We will need to periodically
	// flush the in-memory structure to disk to manage memory usage.
	negativeSentimentFilename string
	// We parse lines in chunks. We may have a partial line we have to reconstruct
	lastStringPart string
	// As we parse lines, we may discover a new URL to fetch.
	nextURL *url.URL
}

func NewMovieSentiment(movie, firstURL string) *MovieSentiment {
	outputFile := fmt.Sprintf("negative_sentiment_%v.txt", movie)
	if err := os.Remove(outputFile); err != nil {
		log.Printf("Warning: could not remove existing negative sentiment file %v: %v",
			outputFile, err)
	}
	return &MovieSentiment{
		movie:                         movie,
		seedURL:                       firstURL,
		negativeSentiment:             []string{},
		maxNegativeSentimentSizeBytes: 1 * 1024,
		negativeSentimentFilename:     outputFile,
	}
}

func (ms *MovieSentiment) Process() error {
	processURL, err := url.Parse(ms.seedURL)
	if err != nil {
		return fmt.Errorf("error parsing seed URL: %v", err)
	}
	for processURL != nil {
		log.Printf("Processing URL for movie %v: %v", ms.movie, processURL.String())
		processURL, err = ms.handleReviewsForEndpoint(*processURL)
		if err != nil {
			return fmt.Errorf("error handling reviews for endpoint: %v", err)
		}
	}
	return nil
}

// BatchReadBytesSize is the amount of bytes we read from the read/closer we get
// via "opening" the URL. NOTE: its set small here for testing purposes.
var BatchReadBytesSize = 1024

// handleReviewsForEndpoint takes a URL and downloads chunks of the reviews file from the URL.
// If the parsing of the reviews file determines another URL needs to be fetched,
// it will return it. If parsing/downloading fails it will return an error indicating
// why.
func (ms *MovieSentiment) handleReviewsForEndpoint(endpointURL url.URL) (*url.URL, error) {
	rc, err := newReadCloserFromURL(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("error opening on URL: %v", err)
	}
	defer rc.Close()

	readCount := 0
	buf := make([]byte, BatchReadBytesSize)
	for {
		bytesRecieved, err := rc.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		log.Printf("Read %v chunked bytes for movie %v [iter: %v]: %v",
			BatchReadBytesSize, ms.movie, readCount+1, bytesRecieved)

		reader := bytes.NewReader(buf[:bytesRecieved])
		if err := ms.parseLines(reader); err != nil {
			return nil, err
		}

		if bytesRecieved < BatchReadBytesSize {
			if err := ms.cleanupPartialLine(); err != nil {
				return nil, err
			}
			break
		}
		readCount++
	}

	var nextURL *url.URL
	if ms.nextURL != nil {
		nextURL = ms.nextURL
		ms.nextURL = nil
	}
	return nextURL, nil
}

func (ms *MovieSentiment) parseLines(buf io.Reader) error {
	scanner := bufio.NewScanner(buf)
	var lastLine string
	if ms.lastStringPart != "" {
		lastLine = ms.lastStringPart
		ms.lastStringPart = ""
	}

	//log.Printf("Starting parse lines for movie %v; starting with last part: %v",
	//	ms.movie, lastLine)

	var i int
	for ; scanner.Scan(); i++ {
		scannerText := scanner.Text()
		if i > 0 {
			ms.analyzeReview(lastLine)
			if len(scannerText) == 0 {
				continue
			}
			lastLine = scannerText
		} else {
			lastLine += scanner.Text()
		}
	}

	// handle last line which may be partial. We only want to analyze/parse
	// full lines. If we end up with a partial line, we have to save it for the next chunk.
	ms.lastStringPart = lastLine

	//log.Printf("Parsed %v lines for movie %v from this chunk; have last part: %v",
	//	i, ms.movie, ms.lastStringPart)
	return nil
}

func (ms *MovieSentiment) cleanupPartialLine() error {
	if ms.lastStringPart == "" {
		return nil
	}

	log.Printf("Cleaning up last partial line for movie %v: %v", ms.movie, ms.lastStringPart)
	if strings.HasPrefix(ms.lastStringPart, "http") || strings.HasPrefix(ms.lastStringPart, "file") {
		parsedURL, err := url.Parse(ms.lastStringPart)
		if err != nil {
			return fmt.Errorf("error parsing next URL: %v", err)
		}
		ms.nextURL = parsedURL
		log.Printf("Discovered next URL for movie %v: %v", ms.movie, ms.nextURL.String())
	} else {
		if err := ms.analyzeReview(ms.lastStringPart); err != nil {
			return fmt.Errorf("error analyzing last partial line: %v", err)
		}
		// Since theres no next URL, we can write any
		// buffered negative sentiment to disk.
		if err := ms.persistNegativeReviews(); err != nil {
			return fmt.Errorf("error persisting negative reviews at end of file: %v", err)
		}
	}

	ms.lastStringPart = ""
	return nil
}

var negativeSentiment = []string{
	"Waste Of time", "Boring", "Terrible acting", "Predictable", "Hated it",
	"Too long", "Bad movie", "Disappointing", "Overrated", "Awful",
}

func (ms *MovieSentiment) analyzeReview(review string) error {
	for _, negativeWord := range negativeSentiment {
		if strings.Contains(strings.ToLower(review), strings.ToLower(negativeWord)) {
			if err := ms.handleNegativeReview(review); err != nil {
				return fmt.Errorf("error appending negative review: %v", err)
			}
		}
	}
	return nil
}

func (ms *MovieSentiment) handleNegativeReview(review string) error {
	ms.negativeSentiment = append(ms.negativeSentiment, review)
	ms.negativeSentimentBytes += len(review)
	if ms.negativeSentimentBytes >= ms.maxNegativeSentimentSizeBytes {
		if err := ms.persistNegativeReviews(); err != nil {
			return fmt.Errorf("error persisting negative reviews: %v", err)
		}
	}
	return nil
}

func (ms *MovieSentiment) persistNegativeReviews() error {
	log.Printf("Persisting %v negative reviews for movie %v to file %v",
		len(ms.negativeSentiment), ms.movie, ms.negativeSentimentFilename)

	if len(ms.negativeSentiment) == 0 {
		return nil
	}

	fd, err := os.OpenFile(ms.negativeSentimentFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening negative sentiment file: %v", err)
	}
	defer fd.Close()

	for _, negReview := range ms.negativeSentiment {
		fd.WriteString(negReview + "\n")
	}

	// reset in-memory structure
	ms.negativeSentiment = []string{}
	ms.negativeSentimentBytes = 0
	return nil
}
