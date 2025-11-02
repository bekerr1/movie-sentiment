package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func randName() string {
	return gofakeit.FirstName() + " " + gofakeit.LastName()
}

func randSentimentReview() string {
	sentiment := rand.Intn(3)

	// Core sentiment phrases
	var core string
	switch sentiment {
	case 0: // positive
		core = gofakeit.RandomString([]string{
			"Loved it", "Amazing movie", "Best film ever", "10/10", "Must watch",
			"Brilliant acting", "Mind-blowing", "Perfect", "Highly recommend", "Outstanding",
		})
	case 1: // negative
		core = gofakeit.RandomString([]string{
			"Waste of time", "Boring", "Terrible acting", "Predictable", "Hated it",
			"Too long", "Bad movie", "Disappointing", "Overrated", "Awful",
		})
	default: // neutral
		core = gofakeit.RandomString([]string{
			"It was okay", "Not bad", "Meh", "Average", "Fine", "Nothing special",
			"Decent", "Mixed feelings", "Alright",
		})
	}

	// Add dynamic filler using gofakeit
	fillers := []string{
		"and the " + gofakeit.Noun() + " was " + gofakeit.Adjective(),
		"but the " + gofakeit.Noun() + " felt " + gofakeit.Adverb(),
		"with " + gofakeit.Adjective() + " " + gofakeit.Noun() + " throughout",
		"especially the " + gofakeit.Noun(),
		"from start to finish",
		"in my opinion",
		"compared to other films",
	}

	numFillers := rand.Intn(3) // 0–2 extra parts
	parts := []string{core}
	for i := 0; i < numFillers; i++ {
		parts = append(parts, fillers[rand.Intn(len(fillers))])
	}

	// Join with natural connectors
	review := strings.Join(parts, ", ")
	review = strings.ReplaceAll(review, " ,", ",")
	review = strings.Title(strings.ToLower(review)) + "."

	return review
}

func main() {
	rand.Seed(time.Now().UnixNano())
	gofakeit.Seed(0)

	file, _ := os.Create("movie_sentiment.txt")
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()

	for i := 0; i < 100; i++ {
		name := randName()
		review := randSentimentReview()
		fmt.Fprintf(w, "%s: %s\n", name, review)
	}
}
