package anki

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnkiConnect_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "version" {
			t.Errorf("expected action 'version', got %s", req.Action)
		}

		resp := ankiResponse{
			Result: float64(6),
			Error:  "",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	if err := ac.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestAnkiConnect_GetDeckNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "deckNames" {
			t.Errorf("expected action 'deckNames', got %s", req.Action)
		}

		resp := ankiResponse{
			Result: []interface{}{"Default", "Japanese", "Programming"},
			Error:  "",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	decks, err := ac.GetDeckNames()
	if err != nil {
		t.Fatalf("GetDeckNames failed: %v", err)
	}

	expected := []string{"Default", "Japanese", "Programming"}
	if len(decks) != len(expected) {
		t.Errorf("expected %d decks, got %d", len(expected), len(decks))
	}

	for i, deck := range decks {
		if deck != expected[i] {
			t.Errorf("expected deck[%d] = %s, got %s", i, expected[i], deck)
		}
	}
}

func TestAnkiConnect_CreateDeck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "createDeck" {
			t.Errorf("expected action 'createDeck', got %s", req.Action)
		}

		params, ok := req.Params.(map[string]interface{})
		if !ok {
			t.Fatal("params is not a map")
		}

		if params["deck"] != "Test Deck" {
			t.Errorf("expected deck name 'Test Deck', got %v", params["deck"])
		}

		resp := ankiResponse{
			Result: float64(1234567890),
			Error:  "",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	if err := ac.CreateDeck("Test Deck"); err != nil {
		t.Errorf("CreateDeck failed: %v", err)
	}
}

func TestAnkiConnect_AddNote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "addNote" {
			t.Errorf("expected action 'addNote', got %s", req.Action)
		}

		resp := ankiResponse{
			Result: float64(1234567890),
			Error:  "",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	note := ankiNote{
		DeckName:  "Test",
		ModelName: "Basic",
		Fields: map[string]string{
			"Front": "Test Front",
			"Back":  "Test Back",
		},
	}

	id, err := ac.AddNote(note)
	if err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	if id != 1234567890 {
		t.Errorf("expected note ID 1234567890, got %d", id)
	}
}

func TestAnkiConnect_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ankiResponse{
			Result: nil,
			Error:  "deck already exists",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	err := ac.CreateDeck("Existing Deck")
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "AnkiConnect error: deck already exists" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeck_PushToAnki(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		callCount++

		var resp ankiResponse
		switch req.Action {
		case "version":
			resp = ankiResponse{Result: float64(6), Error: ""}
		case "createDeck":
			resp = ankiResponse{Result: float64(123), Error: ""}
		case "addNote":
			resp = ankiResponse{Result: float64(456), Error: ""}
		default:
			t.Errorf("unexpected action: %s", req.Action)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	if err := deck.AddCard("Front 1", "Back 1"); err != nil {
		t.Fatalf("Failed to add card 1: %v", err)
	}
	if err := deck.AddCard("Front 2", "Back 2"); err != nil {
		t.Fatalf("Failed to add card 2: %v", err)
	}

	ac := NewAnkiConnectWithURL(server.URL)
	if err := deck.PushToAnki(ac); err != nil {
		t.Errorf("PushToAnki failed: %v", err)
	}

	// Should have called: version, createDeck, addNote x2
	if callCount != 4 {
		t.Errorf("expected 4 API calls, got %d", callCount)
	}
}

func TestAnkiConnect_StoreMediaFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "storeMediaFile" {
			t.Errorf("expected action 'storeMediaFile', got %s", req.Action)
		}

		params, ok := req.Params.(map[string]interface{})
		if !ok {
			t.Fatal("params is not a map")
		}

		if params["filename"] != "test.mp3" {
			t.Errorf("expected filename 'test.mp3', got %v", params["filename"])
		}

		// Check that data is base64 encoded
		if data, ok := params["data"].(string); ok {
			if data != "dGVzdCBhdWRpbyBkYXRh" { // base64 for "test audio data"
				t.Errorf("unexpected base64 data: %s", data)
			}
		}

		resp := ankiResponse{Result: nil, Error: ""}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	testData := []byte("test audio data")
	if err := ac.StoreMediaFile("test.mp3", testData); err != nil {
		t.Errorf("StoreMediaFile failed: %v", err)
	}
}

func TestDeck_PushToAnkiWithMedia(t *testing.T) {
	mediaStored := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		var resp ankiResponse
		switch req.Action {
		case "version":
			resp = ankiResponse{Result: float64(6), Error: ""}
		case "createDeck":
			resp = ankiResponse{Result: float64(123), Error: ""}
		case "storeMediaFile":
			mediaStored = true
			resp = ankiResponse{Result: nil, Error: ""}
		case "addNote":
			resp = ankiResponse{Result: float64(456), Error: ""}
		default:
			t.Errorf("unexpected action: %s", req.Action)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	// Add media
	deck.AddMedia("test.mp3", []byte("audio data"))

	// Add card with audio
	if err := deck.AddCard("Question", "Answer [sound:test.mp3]"); err != nil {
		t.Fatalf("Failed to add card: %v", err)
	}

	ac := NewAnkiConnectWithURL(server.URL)
	if err := deck.PushToAnkiWithMedia(ac, true); err != nil {
		t.Errorf("PushToAnkiWithMedia failed: %v", err)
	}

	if !mediaStored {
		t.Error("Expected media file to be stored")
	}
}

func TestAnkiConnect_GetNotesInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Action != "notesInfo" {
			t.Errorf("expected action 'notesInfo', got %s", req.Action)
		}

		resp := ankiResponse{
			Result: []interface{}{
				map[string]interface{}{
					"noteId": float64(123),
					"fields": map[string]interface{}{
						"Front": map[string]interface{}{
							"value": "Test Front",
							"order": float64(0),
						},
						"Back": map[string]interface{}{
							"value": "Test Back",
							"order": float64(1),
						},
					},
					"tags": []interface{}{"test", "example"},
				},
			},
			Error: "",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	notes, err := ac.GetNotesInfo([]int64{123})
	if err != nil {
		t.Fatalf("GetNotesInfo failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	// Check fields
	fields, ok := notes[0]["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("fields is not a map")
	}

	frontField, ok := fields["Front"].(map[string]interface{})
	if !ok {
		t.Fatal("Front field is not a map")
	}

	if frontField["value"] != "Test Front" {
		t.Errorf("expected Front value 'Test Front', got %v", frontField["value"])
	}
}

func TestDeck_PullFromAnki(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		var resp ankiResponse
		switch req.Action {
		case "version":
			resp = ankiResponse{Result: float64(6), Error: ""}
		case "findNotes":
			resp = ankiResponse{Result: []interface{}{float64(123), float64(456)}, Error: ""}
		case "notesInfo":
			resp = ankiResponse{
				Result: []interface{}{
					map[string]interface{}{
						"noteId": float64(123),
						"fields": map[string]interface{}{
							"Front": map[string]interface{}{"value": "Q1"},
							"Back":  map[string]interface{}{"value": "A1"},
						},
						"tags": []interface{}{"tag1"},
					},
					map[string]interface{}{
						"noteId": float64(456),
						"fields": map[string]interface{}{
							"Front": map[string]interface{}{"value": "Q2"},
							"Back":  map[string]interface{}{"value": "A2"},
						},
						"tags": []interface{}{"tag2"},
					},
				},
				Error: "",
			}
		default:
			t.Errorf("unexpected action: %s", req.Action)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	deck, err := NewDeck("Test Deck")
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}
	defer deck.Close()

	ac := NewAnkiConnectWithURL(server.URL)
	if err := deck.PullFromAnki(ac); err != nil {
		t.Errorf("PullFromAnki failed: %v", err)
	}

	// Verify cards were added
	// Note: We can't easily verify the cards without exposing internal state
	// In a real implementation, we might add a method to count cards
}
