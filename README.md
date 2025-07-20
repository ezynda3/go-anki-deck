# go-anki-deck

A Go package for creating and exporting Anki decks in .apkg format.

## Installation

```bash
go get github.com/ezynda3/go-anki-deck
```

## Usage

### Basic Usage

```go
package main

import (
    "log"
    anki "github.com/ezynda3/go-anki-deck"
)

func main() {
    // Create a new deck
    deck, err := anki.NewDeck("My Deck")
    if err != nil {
        log.Fatal(err)
    }
    defer deck.Close()

    // Add cards
    deck.AddCard("What is the capital of France?", "Paris")
    deck.AddCard("What is 2 + 2?", "4")

    // Add a card with tags
    deck.AddCardWithOptions(
        "What is Go?",
        "A programming language",
        &anki.CardOptions{
            Tags: []string{"programming", "golang"},
        },
    )

    // Save to file
    err = deck.SaveToFile("output.apkg")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Adding Media

```go
// Read an image file
imageData, err := os.ReadFile("image.png")
if err != nil {
    log.Fatal(err)
}

// Add media to deck
deck.AddMedia("image.png", imageData)

// Use the image in a card
deck.AddCard(
    `What is this? <img src="image.png" />`,
    "An example image",
)
```

### Custom Templates

```go
deck, err := anki.NewDeckWithTemplate("Custom Deck", &anki.TemplateOptions{
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
```

## Features

- Create Anki decks programmatically
- Add cards with front and back content
- Support for tags
- Media file support (images, audio, etc.)
- Custom card templates and CSS
- Export to .apkg format compatible with Anki

## API Reference

### Types

#### `Deck`
The main type representing an Anki deck.

#### `CardOptions`
Options for adding cards:
- `Tags []string` - Tags to associate with the card

#### `TemplateOptions`
Options for customizing card templates:
- `QuestionFormat string` - HTML template for the question side
- `AnswerFormat string` - HTML template for the answer side
- `CSS string` - CSS styles for the cards

### Functions

#### `NewDeck(name string) (*Deck, error)`
Creates a new deck with the default template.

#### `NewDeckWithTemplate(name string, opts *TemplateOptions) (*Deck, error)`
Creates a new deck with a custom template.

#### `(*Deck) AddCard(front, back string) error`
Adds a card to the deck.

#### `(*Deck) AddCardWithOptions(front, back string, opts *CardOptions) error`
Adds a card with additional options like tags.

#### `(*Deck) AddMedia(filename string, data []byte)`
Adds a media file to the deck.

#### `(*Deck) Save() ([]byte, error)`
Exports the deck as .apkg format and returns the data.

#### `(*Deck) SaveToFile(filename string) error`
Exports the deck directly to a file.

#### `(*Deck) Close() error`
Closes the deck and releases resources.

## License

MIT

## Credits

This package is a Go port of the JavaScript [anki-apkg-export](https://github.com/repeat-space/anki-apkg-export) library.