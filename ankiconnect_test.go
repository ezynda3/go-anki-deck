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
