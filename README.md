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

### Adding Audio

```go
// Method 1: Using AddAudio helper
audioData, err := os.ReadFile("pronunciation.mp3")
if err != nil {
    log.Fatal(err)
}
soundTag := deck.AddAudio("pronunciation.mp3", audioData)
deck.AddCard("How do you pronounce 'hello'?", "Hello " + soundTag)

// Method 2: Using AddCardWithAudio convenience method
deck.AddCardWithAudio(
    "What sound is this?",
    "A bell ringing",
    "bell.mp3",
    audioData,
)

// Method 3: Using CardOptions for audio on both sides
deck.AddMedia("question.mp3", questionAudio)
deck.AddMedia("answer.mp3", answerAudio)
deck.AddCardWithOptions(
    "Listen to the question",
    "Here's the answer",
    &anki.CardOptions{
        Tags:       []string{"audio", "listening"},
        FrontAudio: "question.mp3",
        BackAudio:  "answer.mp3",
    },
)
```

### Adding Images

```go
// Method 1: Using AddImage helper
imageData, err := os.ReadFile("diagram.png")
if err != nil {
    log.Fatal(err)
}
imgTag := deck.AddImage("diagram.png", imageData)
deck.AddCard("What does this show?", "A diagram " + imgTag)

// Method 2: Using AddCardWithImage convenience method
deck.AddCardWithImage(
    "Identify this structure",
    "The Eiffel Tower",
    "tower.jpg",
    imageData,
)

// Method 3: Using CardOptions for images
deck.AddCardWithOptions(
    "Compare these images",
    "They show before and after",
    &anki.CardOptions{
        FrontImage: "before.png",
        BackImage:  "after.png",
    },
)
```

### Adding Videos

```go
// Method 1: Using AddVideo helper
videoData, err := os.ReadFile("demo.mp4")
if err != nil {
    log.Fatal(err)
}
videoTag := deck.AddVideo("demo.mp4", videoData)
deck.AddCard("Watch this technique", "Explanation: " + videoTag)

// Method 2: Using AddCardWithVideo convenience method
deck.AddCardWithVideo(
    "What's happening in this video?",
    "A chemical reaction",
    "reaction.webm",
    videoData,
)

// Method 3: Multimedia card with all media types
deck.AddCardWithOptions(
    "Study this content",
    "Complete explanation",
    &anki.CardOptions{
        Tags:       []string{"multimedia"},
        FrontImage: "slide.png",
        FrontAudio: "narration.mp3",
        BackVideo:  "animation.mp4",
    },
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

### AnkiConnect Integration

This package supports syncing decks directly to Anki desktop using the [AnkiConnect](https://ankiweb.net/shared/info/2055492159) addon.

#### Prerequisites

1. Install [Anki](https://apps.ankiweb.net/) desktop application
2. Install the [AnkiConnect](https://ankiweb.net/shared/info/2055492159) addon in Anki
3. Ensure Anki is running when using sync features

#### Basic Usage

```go
// Create and populate a deck
deck, err := anki.NewDeck("My Deck")
if err != nil {
    log.Fatal(err)
}
defer deck.Close()

deck.AddCard("Question", "Answer")

// Create AnkiConnect client
ac := anki.NewAnkiConnect()

// Push deck to Anki
err = deck.PushToAnki(ac)
if err != nil {
    log.Fatal(err)
}
```

#### Advanced Sync Options

```go
// Sync with options
opts := &anki.SyncOptions{
    UpdateExisting: true,  // Update existing cards instead of skipping
    DeleteMissing: false,  // Don't delete cards not in local deck
    SyncMedia: true,      // Sync media files (images, audio, video)
}

err := deck.SyncToAnki(ac, opts)
```

#### Media Sync

```go
// Push deck with media files
err := deck.PushToAnkiWithMedia(ac, true)

// Media files are automatically detected from card content
// Supported formats:
// - Audio: [sound:filename.mp3]
// - Images: <img src="filename.png">
// - Videos: <video src="filename.mp4">
```

#### Bidirectional Sync

```go
// Pull cards from Anki to local deck
err := deck.PullFromAnki(ac)

// This will:
// 1. Find all cards in the Anki deck
// 2. Clear local cards
// 3. Import cards from Anki with their tags
```

#### AnkiConnect Operations

```go
// Check connection
err := ac.Ping()

// Get all deck names
decks, err := ac.GetDeckNames()

// Create a deck directly
err := ac.CreateDeck("New Deck")

// Delete a deck
err := ac.DeleteDeck("Old Deck")

// Trigger sync to AnkiWeb
err := ac.Sync()
```

#### Custom AnkiConnect URL

If AnkiConnect is running on a different port or host:

```go
ac := anki.NewAnkiConnectWithURL("http://localhost:8765")
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
- `FrontAudio string` - Audio filename to play on the front of the card
- `BackAudio string` - Audio filename to play on the back of the card
- `FrontImage string` - Image filename to display on the front of the card
- `BackImage string` - Image filename to display on the back of the card
- `FrontVideo string` - Video filename to display on the front of the card
- `BackVideo string` - Video filename to display on the back of the card

#### `TemplateOptions`
Options for customizing card templates:
- `QuestionFormat string` - HTML template for the question side
- `AnswerFormat string` - HTML template for the answer side
- `CSS string` - CSS styles for the cards

#### `AnkiConnect`
Client for communicating with AnkiConnect addon:
- `URL string` - AnkiConnect server URL (default: http://localhost:8765)
- `Version int` - AnkiConnect API version (default: 6)

#### `SyncOptions`
Options for deck synchronization:
- `UpdateExisting bool` - Update existing cards
- `DeleteMissing bool` - Delete cards not in local deck
- `SyncMedia bool` - Sync media files

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

#### `(*Deck) AddAudio(filename string, data []byte) string`
Adds an audio file to the deck and returns the Anki sound tag.

#### `(*Deck) AddImage(filename string, data []byte) string`
Adds an image file to the deck and returns the HTML img tag.

#### `(*Deck) AddVideo(filename string, data []byte) string`
Adds a video file to the deck and returns the HTML video tag.

#### `(*Deck) AddCardWithAudio(front, back, audioFile string, audioData []byte) error`
Adds a card with an audio file attached to the back.

#### `(*Deck) AddCardWithImage(front, back, imageFile string, imageData []byte) error`
Adds a card with an image file attached to the back.

#### `(*Deck) AddCardWithVideo(front, back, videoFile string, videoData []byte) error`
Adds a card with a video file attached to the back.

#### `(*Deck) Save() ([]byte, error)`
Exports the deck as .apkg format and returns the data.

#### `(*Deck) SaveToFile(filename string) error`
Exports the deck directly to a file.

#### `(*Deck) Close() error`
Closes the deck and releases resources.

#### `(*Deck) PushToAnki(client *AnkiConnect) error`
Pushes the entire deck to Anki, creating it if necessary.

#### `(*Deck) SyncToAnki(client *AnkiConnect, opts *SyncOptions) error`
Performs a more sophisticated sync with options.

### AnkiConnect Functions

#### `NewAnkiConnect() *AnkiConnect`
Creates a new AnkiConnect client with default settings.

#### `NewAnkiConnectWithURL(url string) *AnkiConnect`
Creates a new AnkiConnect client with custom URL.

#### `(*AnkiConnect) Ping() error`
Checks if AnkiConnect is available.

#### `(*AnkiConnect) GetDeckNames() ([]string, error)`
Returns all deck names in Anki.

#### `(*AnkiConnect) CreateDeck(name string) error`
Creates a new deck in Anki.

#### `(*AnkiConnect) DeleteDeck(name string) error`
Deletes a deck and all its cards.

#### `(*AnkiConnect) Sync() error`
Triggers Anki to sync with AnkiWeb.

#### `(*AnkiConnect) StoreMediaFile(filename string, data []byte) error`
Stores a media file in Anki's media folder with base64 encoding.

#### `(*AnkiConnect) GetNotesInfo(noteIDs []int64) ([]map[string]interface{}, error)`
Retrieves detailed information about notes.

#### `(*Deck) PushToAnkiWithMedia(client *AnkiConnect, syncMedia bool) error`
Pushes the deck to Anki with optional media sync.

#### `(*Deck) PullFromAnki(client *AnkiConnect) error`
Pulls cards from Anki deck and updates the local deck.

## License

MIT

## Credits

This package is a Go port of the JavaScript [anki-apkg-export](https://github.com/repeat-space/anki-apkg-export) library.