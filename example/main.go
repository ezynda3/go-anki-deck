package main

import (
	"fmt"
	"log"
	"os"

	anki "github.com/ezynda3/go-anki-deck"
)

func main() {
	// Create a new deck
	deck, err := anki.NewDeck("My Go Anki Deck")
	if err != nil {
		log.Fatal(err)
	}
	defer deck.Close()

	// Add some cards
	err = deck.AddCard("What is the capital of France?", "Paris")
	if err != nil {
		log.Fatal(err)
	}

	err = deck.AddCard("What is 2 + 2?", "4")
	if err != nil {
		log.Fatal(err)
	}

	// Add a card with tags
	err = deck.AddCardWithOptions(
		"What is the Go programming language?",
		"A statically typed, compiled programming language designed at Google",
		&anki.CardOptions{
			Tags: []string{"programming", "golang", "computer science"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add media (example with a simple PNG)
	if imageData, err := os.ReadFile("gopher.png"); err == nil {
		deck.AddMedia("gopher.png", imageData)

		// Add a card that uses the image
		err = deck.AddCard(
			`What is the Go mascot? <img src="gopher.png" />`,
			"The Go Gopher",
		)
		if err != nil {
			log.Printf("Failed to add card with image: %v", err)
		}
	}

	// Audio examples
	// Example 1: Using AddAudio helper
	if audioData, err := os.ReadFile("pronunciation.mp3"); err == nil {
		soundTag := deck.AddAudio("pronunciation.mp3", audioData)
		err = deck.AddCard(
			"How do you pronounce 'hello'?",
			"Hello "+soundTag,
		)
		if err != nil {
			log.Printf("Failed to add card with audio: %v", err)
		}
	}

	// Example 2: Using AddCardWithAudio convenience method
	if audioData, err := os.ReadFile("word.mp3"); err == nil {
		err = deck.AddCardWithAudio(
			"What word is this?",
			"Example",
			"word.mp3",
			audioData,
		)
		if err != nil {
			log.Printf("Failed to add card with audio: %v", err)
		}
	}

	// Example 3: Using CardOptions with audio on both sides
	if frontAudio, err := os.ReadFile("question.mp3"); err == nil {
		if backAudio, err := os.ReadFile("answer.mp3"); err == nil {
			deck.AddMedia("question.mp3", frontAudio)
			deck.AddMedia("answer.mp3", backAudio)

			err = deck.AddCardWithOptions(
				"Listen to the question",
				"Here's the answer",
				&anki.CardOptions{
					Tags:       []string{"audio", "example"},
					FrontAudio: "question.mp3",
					BackAudio:  "answer.mp3",
				},
			)
			if err != nil {
				log.Printf("Failed to add card with audio options: %v", err)
			}
		}
	}

	// Image examples
	// Example 1: Using AddImage helper
	if imageData, err := os.ReadFile("diagram.png"); err == nil {
		imgTag := deck.AddImage("diagram.png", imageData)
		err = deck.AddCard(
			"What does this diagram show?",
			"A flowchart "+imgTag,
		)
		if err != nil {
			log.Printf("Failed to add card with image: %v", err)
		}
	}

	// Example 2: Using AddCardWithImage convenience method
	if imageData, err := os.ReadFile("chart.jpg"); err == nil {
		err = deck.AddCardWithImage(
			"Analyze this chart",
			"This shows quarterly revenue growth",
			"chart.jpg",
			imageData,
		)
		if err != nil {
			log.Printf("Failed to add card with image: %v", err)
		}
	}

	// Video examples
	// Example 1: Using AddVideo helper
	if videoData, err := os.ReadFile("demo.mp4"); err == nil {
		videoTag := deck.AddVideo("demo.mp4", videoData)
		err = deck.AddCard(
			"Watch this demonstration",
			"The video shows how to use the tool "+videoTag,
		)
		if err != nil {
			log.Printf("Failed to add card with video: %v", err)
		}
	}

	// Example 2: Using AddCardWithVideo convenience method
	if videoData, err := os.ReadFile("tutorial.webm"); err == nil {
		err = deck.AddCardWithVideo(
			"What technique is demonstrated?",
			"The proper form for a deadlift",
			"tutorial.webm",
			videoData,
		)
		if err != nil {
			log.Printf("Failed to add card with video: %v", err)
		}
	}

	// Example 3: Multimedia card with all media types
	if audioData, err := os.ReadFile("narration.mp3"); err == nil {
		if imageData, err := os.ReadFile("slide.png"); err == nil {
			if videoData, err := os.ReadFile("animation.mp4"); err == nil {
				deck.AddMedia("narration.mp3", audioData)
				deck.AddMedia("slide.png", imageData)
				deck.AddMedia("animation.mp4", videoData)

				err = deck.AddCardWithOptions(
					"Study this multimedia content",
					"This demonstrates the water cycle",
					&anki.CardOptions{
						Tags:       []string{"science", "multimedia"},
						FrontImage: "slide.png",
						FrontAudio: "narration.mp3",
						BackVideo:  "animation.mp4",
					},
				)
				if err != nil {
					log.Printf("Failed to add multimedia card: %v", err)
				}
			}
		}
	}

	// Save the deck
	apkgData, err := deck.Save()
	if err != nil {
		log.Fatal(err)
	}

	// Write to file
	err = os.WriteFile("output.apkg", apkgData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deck exported successfully to output.apkg")

	// Alternative: Save directly to file
	deck2, err := anki.NewDeck("Another Deck")
	if err != nil {
		log.Fatal(err)
	}
	defer deck2.Close()

	err = deck2.AddCard("Hello", "World")
	if err != nil {
		log.Fatal(err)
	}

	err = deck2.SaveToFile("another.apkg")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Second deck exported successfully to another.apkg")
}
