package anki

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	if deck.name != "Test Deck" {
		t.Errorf("Expected deck name 'Test Deck', got '%s'", deck.name)
	}
}

func TestAddCard(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	err = deck.AddCard("Front", "Back")
	if err != nil {
		t.Errorf("Failed to add card: %v", err)
	}

	// Verify card was added
	var count int
	err = deck.db.QueryRow("SELECT COUNT(*) FROM cards").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query cards: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 card, got %d", count)
	}

	// Verify note was added
	err = deck.db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query notes: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 note, got %d", count)
	}
}

func TestAddCardWithTags(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	tags := []string{"tag1", "tag2", "multi word tag"}
	err = deck.AddCardWithOptions("Front", "Back", &CardOptions{
		Tags: tags,
	})
	if err != nil {
		t.Errorf("Failed to add card with tags: %v", err)
	}

	// Verify tags were stored correctly
	var storedTags string
	err = deck.db.QueryRow("SELECT tags FROM notes").Scan(&storedTags)
	if err != nil {
		t.Errorf("Failed to query tags: %v", err)
	}

	expectedTags := " tag1 tag2 multi_word_tag "
	if storedTags != expectedTags {
		t.Errorf("Expected tags '%s', got '%s'", expectedTags, storedTags)
	}
}

func TestAddMedia(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	mediaData := []byte("test media content")
	deck.AddMedia("test.txt", mediaData)

	if len(deck.media) != 1 {
		t.Errorf("Expected 1 media file, got %d", len(deck.media))
	}

	if deck.media[0].Filename != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got '%s'", deck.media[0].Filename)
	}

	if !bytes.Equal(deck.media[0].Data, mediaData) {
		t.Errorf("Media data mismatch")
	}
}

func TestSave(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Add a card
	err = deck.AddCard("Question", "Answer")
	if err != nil {
		t.Fatalf("Failed to add card: %v", err)
	}

	// Add media
	deck.AddMedia("test.txt", []byte("test content"))

	// Save deck
	data, err := deck.Save()
	if err != nil {
		t.Fatalf("Failed to save deck: %v", err)
	}

	// Verify it's a valid ZIP file
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	// Check required files exist
	requiredFiles := map[string]bool{
		"collection.anki2": false,
		"media":            false,
		"0":                false, // First media file
	}

	for _, file := range reader.File {
		if _, ok := requiredFiles[file.Name]; ok {
			requiredFiles[file.Name] = true
		}
	}

	for name, found := range requiredFiles {
		if !found {
			t.Errorf("Required file '%s' not found in ZIP", name)
		}
	}

	// Verify media manifest
	for _, file := range reader.File {
		if file.Name == "media" {
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("Failed to open media file: %v", err)
			}
			defer rc.Close()

			var buf bytes.Buffer
			_, err = buf.ReadFrom(rc)
			if err != nil {
				t.Fatalf("Failed to read media file: %v", err)
			}

			var mediaMap map[string]string
			err = json.Unmarshal(buf.Bytes(), &mediaMap)
			if err != nil {
				t.Fatalf("Failed to parse media manifest: %v", err)
			}

			if mediaMap["0"] != "test.txt" {
				t.Errorf("Expected media file '0' to be 'test.txt', got '%s'", mediaMap["0"])
			}
			break
		}
	}
}

func TestCustomTemplate(t *testing.T) {
	customCSS := ".card { color: red; }"
	customQuestion := "<b>{{Front}}</b>"
	customAnswer := "{{Back}}"

	deck, err := NewDeckWithTemplate("Custom Deck", &TemplateOptions{
		QuestionFormat: customQuestion,
		AnswerFormat:   customAnswer,
		CSS:            customCSS,
	})
	if err != nil {
		t.Fatalf("Failed to create deck with custom template: %v", err)
	}
	defer deck.Close()

	// Verify template was applied
	var modelsJSON string
	err = deck.db.QueryRow("SELECT models FROM col WHERE id = 1").Scan(&modelsJSON)
	if err != nil {
		t.Fatalf("Failed to query models: %v", err)
	}

	var models map[string]interface{}
	err = json.Unmarshal([]byte(modelsJSON), &models)
	if err != nil {
		t.Fatalf("Failed to parse models: %v", err)
	}

	// Check if custom template values are present
	for _, model := range models {
		m := model.(map[string]interface{})
		if css, ok := m["css"].(string); ok && css == customCSS {
			// Found our custom CSS
			tmpls := m["tmpls"].([]interface{})
			if len(tmpls) > 0 {
				tmpl := tmpls[0].(map[string]interface{})
				if tmpl["qfmt"] != customQuestion {
					t.Errorf("Expected question format '%s', got '%s'", customQuestion, tmpl["qfmt"])
				}
				if tmpl["afmt"] != customAnswer {
					t.Errorf("Expected answer format '%s', got '%s'", customAnswer, tmpl["afmt"])
				}
			}
			return
		}
	}

	t.Error("Custom template not found in models")
}

func TestDuplicateCard(t *testing.T) {
	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Add the same card twice
	err = deck.AddCard("Same Front", "Same Back")
	if err != nil {
		t.Errorf("Failed to add first card: %v", err)
	}

	err = deck.AddCard("Same Front", "Same Back")
	if err != nil {
		t.Errorf("Failed to add duplicate card: %v", err)
	}

	// Should still have only one note (duplicates update existing)
	var noteCount int
	err = deck.db.QueryRow("SELECT COUNT(DISTINCT guid) FROM notes").Scan(&noteCount)
	if err != nil {
		t.Errorf("Failed to query notes: %v", err)
	}
	if noteCount != 1 {
		t.Errorf("Expected 1 unique note, got %d", noteCount)
	}
}

func TestAddAudio(t *testing.T) {
	deck, err := NewDeck("Audio Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Test AddAudio helper
	audioData := []byte("fake audio data")
	soundTag := deck.AddAudio("test.mp3", audioData)
	if soundTag != "[sound:test.mp3]" {
		t.Errorf("Expected '[sound:test.mp3]', got '%s'", soundTag)
	}

	// Verify media was added
	if len(deck.media) != 1 {
		t.Errorf("Expected 1 media file, got %d", len(deck.media))
	}
	if deck.media[0].Filename != "test.mp3" {
		t.Errorf("Expected filename 'test.mp3', got '%s'", deck.media[0].Filename)
	}
}

func TestAddCardWithAudio(t *testing.T) {
	deck, err := NewDeck("Audio Card Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Test AddCardWithAudio
	audioData := []byte("fake audio data")
	err = deck.AddCardWithAudio("What sound is this?", "A test sound", "test.mp3", audioData)
	if err != nil {
		t.Errorf("Failed to add card with audio: %v", err)
	}

	// Verify media was added
	if len(deck.media) != 1 {
		t.Errorf("Expected 1 media file, got %d", len(deck.media))
	}

	// Verify card was created with audio tag
	var flds string
	err = deck.db.QueryRow("SELECT flds FROM notes").Scan(&flds)
	if err != nil {
		t.Fatalf("Failed to query note fields: %v", err)
	}

	if !strings.Contains(flds, "[sound:test.mp3]") {
		t.Errorf("Expected fields to contain '[sound:test.mp3]', got '%s'", flds)
	}
}

func TestAddCardWithOptions_Audio(t *testing.T) {
	deck, err := NewDeck("Audio Options Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Add audio files first
	frontAudio := []byte("front audio data")
	backAudio := []byte("back audio data")
	deck.AddMedia("front.mp3", frontAudio)
	deck.AddMedia("back.mp3", backAudio)

	// Add card with audio options
	err = deck.AddCardWithOptions(
		"Question",
		"Answer",
		&CardOptions{
			Tags:       []string{"audio", "test"},
			FrontAudio: "front.mp3",
			BackAudio:  "back.mp3",
		},
	)
	if err != nil {
		t.Errorf("Failed to add card with audio options: %v", err)
	}

	// Verify card fields contain audio tags
	var flds string
	err = deck.db.QueryRow("SELECT flds FROM notes").Scan(&flds)
	if err != nil {
		t.Fatalf("Failed to query note fields: %v", err)
	}

	parts := strings.Split(flds, separator)
	if len(parts) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(parts))
	}

	if !strings.Contains(parts[0], "[sound:front.mp3]") {
		t.Errorf("Expected front to contain '[sound:front.mp3]', got '%s'", parts[0])
	}
	if !strings.Contains(parts[1], "[sound:back.mp3]") {
		t.Errorf("Expected back to contain '[sound:back.mp3]', got '%s'", parts[1])
	}

	// Verify tags
	var tags string
	err = deck.db.QueryRow("SELECT tags FROM notes").Scan(&tags)
	if err != nil {
		t.Fatalf("Failed to query tags: %v", err)
	}
	if !strings.Contains(tags, "audio") || !strings.Contains(tags, "test") {
		t.Errorf("Expected tags to contain 'audio' and 'test', got '%s'", tags)
	}
}

func BenchmarkAddCard(b *testing.B) {
	deck, err := NewDeck("Benchmark Deck")
	if err != nil {
		b.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := deck.AddCard(
			fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i),
		)
		if err != nil {
			b.Fatalf("Failed to add card: %v", err)
		}
	}
}

func BenchmarkSave(b *testing.B) {
	deck, err := NewDeck("Benchmark Deck")
	if err != nil {
		b.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Add some cards
	for i := 0; i < 100; i++ {
		err := deck.AddCard(
			fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i),
		)
		if err != nil {
			b.Fatalf("Failed to add card: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := deck.Save()
		if err != nil {
			b.Fatalf("Failed to save: %v", err)
		}
	}
}
