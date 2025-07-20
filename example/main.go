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

	deck2.AddCard("Hello", "World")

	err = deck2.SaveToFile("another.apkg")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Second deck exported successfully to another.apkg")
}

// Example with custom template
func exampleWithCustomTemplate() {
	deck, err := anki.NewDeckWithTemplate("Custom Template Deck", &anki.TemplateOptions{
		QuestionFormat: `<div class="question">{{Front}}</div>`,
		AnswerFormat: `{{FrontSide}}
<hr id="answer">
<div class="answer">{{Back}}</div>`,
		CSS: `.card {
	font-family: Georgia, serif;
	font-size: 24px;
	text-align: center;
	color: #333;
	background-color: #f5f5f5;
}
.question {
	color: blue;
	font-weight: bold;
}
.answer {
	color: green;
	margin-top: 20px;
}`,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer deck.Close()

	deck.AddCard("Custom styled question", "Custom styled answer")
	deck.SaveToFile("custom_template.apkg")
}
