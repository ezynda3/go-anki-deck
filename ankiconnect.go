package anki

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultAnkiConnectURL = "http://localhost:8765"
	ankiConnectVersion    = 6
)

// AnkiConnect represents a client for communicating with AnkiConnect addon
type AnkiConnect struct {
	URL     string
	Version int
	client  *http.Client
}

// SyncOptions controls the behavior of deck synchronization
type SyncOptions struct {
	UpdateExisting bool // Update existing cards
	DeleteMissing  bool // Delete cards not in local deck
	SyncMedia      bool // Sync media files
}

// ankiRequest represents a request to AnkiConnect API
type ankiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// ankiResponse represents a response from AnkiConnect API
type ankiResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// NewAnkiConnect creates a new AnkiConnect client with default settings
func NewAnkiConnect() *AnkiConnect {
	return &AnkiConnect{
		URL:     defaultAnkiConnectURL,
		Version: ankiConnectVersion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAnkiConnectWithURL creates a new AnkiConnect client with custom URL
func NewAnkiConnectWithURL(url string) *AnkiConnect {
	ac := NewAnkiConnect()
	ac.URL = url
	return ac
}

// invoke makes a request to AnkiConnect API
func (ac *AnkiConnect) invoke(action string, params interface{}) (interface{}, error) {
	req := ankiRequest{
		Action:  action,
		Version: ac.Version,
		Params:  params,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := ac.client.Post(ac.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ankiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("AnkiConnect error: %s", result.Error)
	}

	return result.Result, nil
}

// Ping checks if AnkiConnect is available
func (ac *AnkiConnect) Ping() error {
	_, err := ac.invoke("version", nil)
	return err
}

// GetDeckNames returns all deck names in Anki
func (ac *AnkiConnect) GetDeckNames() ([]string, error) {
	result, err := ac.invoke("deckNames", nil)
	if err != nil {
		return nil, err
	}

	names, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	deckNames := make([]string, len(names))
	for i, name := range names {
		deckNames[i], ok = name.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected deck name type")
		}
	}

	return deckNames, nil
}

// CreateDeck creates a new deck in Anki
func (ac *AnkiConnect) CreateDeck(name string) error {
	params := map[string]string{"deck": name}
	_, err := ac.invoke("createDeck", params)
	return err
}

// DeleteDeck deletes a deck and all its cards
func (ac *AnkiConnect) DeleteDeck(name string) error {
	params := map[string]interface{}{
		"decks":    []string{name},
		"cardsToo": true,
	}
	_, err := ac.invoke("deleteDecks", params)
	return err
}

// ankiNote represents a note in AnkiConnect format
type ankiNote struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Fields    map[string]string      `json:"fields"`
	Tags      []string               `json:"tags,omitempty"`
	Audio     []ankiMedia            `json:"audio,omitempty"`
	Picture   []ankiMedia            `json:"picture,omitempty"`
	Video     []ankiMedia            `json:"video,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// ankiMedia represents media attachment in AnkiConnect format
type ankiMedia struct {
	Path     string   `json:"path,omitempty"`
	Filename string   `json:"filename,omitempty"`
	Fields   []string `json:"fields,omitempty"`
	Data     string   `json:"data,omitempty"`
}

// AddNote adds a single note to Anki
func (ac *AnkiConnect) AddNote(note ankiNote) (int64, error) {
	params := map[string]interface{}{"note": note}
	result, err := ac.invoke("addNote", params)
	if err != nil {
		return 0, err
	}

	// AnkiConnect returns note ID as float64
	if id, ok := result.(float64); ok {
		return int64(id), nil
	}

	return 0, fmt.Errorf("unexpected note ID type")
}

// FindNotes searches for notes matching a query
func (ac *AnkiConnect) FindNotes(query string) ([]int64, error) {
	params := map[string]string{"query": query}
	result, err := ac.invoke("findNotes", params)
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	noteIDs := make([]int64, len(ids))
	for i, id := range ids {
		if fid, ok := id.(float64); ok {
			noteIDs[i] = int64(fid)
		} else {
			return nil, fmt.Errorf("unexpected note ID type")
		}
	}

	return noteIDs, nil
}

// UpdateNoteFields updates fields of an existing note
func (ac *AnkiConnect) UpdateNoteFields(noteID int64, fields map[string]string) error {
	params := map[string]interface{}{
		"note": map[string]interface{}{
			"id":     noteID,
			"fields": fields,
		},
	}
	_, err := ac.invoke("updateNoteFields", params)
	return err
}

// StoreMediaFile stores a media file in Anki's media folder
func (ac *AnkiConnect) StoreMediaFile(filename string, data []byte) error {
	// AnkiConnect expects base64 encoded data
	encodedData := base64.StdEncoding.EncodeToString(data)
	params := map[string]interface{}{
		"filename": filename,
		"data":     encodedData,
	}
	_, err := ac.invoke("storeMediaFile", params)
	return err
}

// Sync triggers Anki to sync with AnkiWeb
func (ac *AnkiConnect) Sync() error {
	_, err := ac.invoke("sync", nil)
	return err
}

// GetNotesInfo retrieves detailed information about notes
func (ac *AnkiConnect) GetNotesInfo(noteIDs []int64) ([]map[string]interface{}, error) {
	params := map[string]interface{}{"notes": noteIDs}
	result, err := ac.invoke("notesInfo", params)
	if err != nil {
		return nil, err
	}

	notes, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	notesInfo := make([]map[string]interface{}, len(notes))
	for i, note := range notes {
		noteMap, ok := note.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected note type")
		}
		notesInfo[i] = noteMap
	}

	return notesInfo, nil
}

// PullFromAnki pulls cards from Anki deck and updates the local deck
func (d *Deck) PullFromAnki(client *AnkiConnect) error {
	// Check connection
	if err := client.Ping(); err != nil {
		return fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}

	// Find notes in the deck
	query := fmt.Sprintf("deck:\"%s\"", d.name)
	noteIDs, err := client.FindNotes(query)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	if len(noteIDs) == 0 {
		return nil // No notes to pull
	}

	// Get detailed note information
	notesInfo, err := client.GetNotesInfo(noteIDs)
	if err != nil {
		return fmt.Errorf("failed to get notes info: %w", err)
	}

	// Clear existing cards in the deck
	// Note: In a production implementation, you might want to merge instead
	if _, err := d.db.Exec("DELETE FROM cards WHERE did = ?", d.topDeckID); err != nil {
		return fmt.Errorf("failed to clear existing cards: %w", err)
	}
	if _, err := d.db.Exec("DELETE FROM notes"); err != nil {
		return fmt.Errorf("failed to clear existing notes: %w", err)
	}

	// Add each note from Anki
	for _, noteInfo := range notesInfo {
		fields, ok := noteInfo["fields"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract front and back fields
		var front, back string
		if frontField, ok := fields["Front"].(map[string]interface{}); ok {
			if value, ok := frontField["value"].(string); ok {
				front = value
			}
		}
		if backField, ok := fields["Back"].(map[string]interface{}); ok {
			if value, ok := backField["value"].(string); ok {
				back = value
			}
		}

		// Extract tags
		var tags []string
		if tagsInterface, ok := noteInfo["tags"].([]interface{}); ok {
			for _, tag := range tagsInterface {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

		// Add the card to local deck
		opts := &CardOptions{
			Tags: tags,
		}
		if err := d.AddCardWithOptions(front, back, opts); err != nil {
			return fmt.Errorf("failed to add card: %w", err)
		}
	}

	return nil
}

// syncWithExisting syncs the deck with existing notes in Anki
func (d *Deck) syncWithExisting(client *AnkiConnect, existingMap map[string]int64, syncMedia bool) error {
	// Sync media files first if requested
	if syncMedia && len(d.media) > 0 {
		for _, media := range d.media {
			if err := client.StoreMediaFile(media.Filename, media.Data); err != nil {
				fmt.Printf("Warning: failed to sync media file %s: %v\n", media.Filename, err)
			}
		}
	}

	// Query cards from the database
	rows, err := d.db.Query(`
		SELECT n.flds, n.tags 
		FROM notes n 
		JOIN cards c ON c.nid = n.id 
		WHERE c.did = ?`, d.topDeckID)
	if err != nil {
		return fmt.Errorf("failed to query cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Process each card
	for rows.Next() {
		var flds, tags string
		if err := rows.Scan(&flds, &tags); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Split fields (front and back)
		fields := strings.Split(flds, separator)
		if len(fields) < 2 {
			continue
		}

		key := fields[0] + "|" + fields[1]

		// Check if note already exists
		if noteID, exists := existingMap[key]; exists {
			// Update existing note
			updateFields := map[string]string{
				"Front": fields[0],
				"Back":  fields[1],
			}

			if err := client.UpdateNoteFields(noteID, updateFields); err != nil {
				return fmt.Errorf("failed to update note %d: %w", noteID, err)
			}
		} else {
			// Add new note
			note := ankiNote{
				DeckName:  d.name,
				ModelName: "Basic",
				Fields: map[string]string{
					"Front": fields[0],
					"Back":  fields[1],
				},
				Options: map[string]interface{}{
					"allowDuplicate": false,
				},
			}

			// Parse tags if present
			if tags != "" {
				note.Tags = strings.Fields(tags)
			}

			// Extract media references if syncMedia is enabled
			if syncMedia {
				note.Audio = extractMediaReferences(fields[0], fields[1], "sound")
				note.Picture = extractMediaReferences(fields[0], fields[1], "img")
				note.Video = extractMediaReferences(fields[0], fields[1], "video")
			}

			if _, err := client.AddNote(note); err != nil {
				if err.Error() != "AnkiConnect error: cannot create note because it is a duplicate" {
					return fmt.Errorf("failed to add card: %w", err)
				}
			}
		}
	}

	return rows.Err()
}

// PushToAnki pushes the entire deck to Anki, creating it if necessary
func (d *Deck) PushToAnki(client *AnkiConnect) error {
	return d.PushToAnkiWithMedia(client, false)
}

// PushToAnkiWithMedia pushes the deck to Anki with optional media sync
func (d *Deck) PushToAnkiWithMedia(client *AnkiConnect, syncMedia bool) error {
	// Check connection
	if err := client.Ping(); err != nil {
		return fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}

	// Create deck if it doesn't exist
	if err := client.CreateDeck(d.name); err != nil {
		// Ignore error if deck already exists
		if err.Error() != "AnkiConnect error: deck already exists" {
			return fmt.Errorf("failed to create deck: %w", err)
		}
	}

	// Sync media files first if requested
	if syncMedia && len(d.media) > 0 {
		for _, media := range d.media {
			if err := client.StoreMediaFile(media.Filename, media.Data); err != nil {
				// Log but don't fail on media errors
				fmt.Printf("Warning: failed to sync media file %s: %v\n", media.Filename, err)
			}
		}
	}

	// Query cards from the database
	rows, err := d.db.Query(`
		SELECT n.flds, n.tags 
		FROM notes n 
		JOIN cards c ON c.nid = n.id 
		WHERE c.did = ?`, d.topDeckID)
	if err != nil {
		return fmt.Errorf("failed to query cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Add each card
	for rows.Next() {
		var flds, tags string
		if err := rows.Scan(&flds, &tags); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Split fields (front and back)
		fields := strings.Split(flds, separator)
		if len(fields) < 2 {
			continue
		}

		note := ankiNote{
			DeckName:  d.name,
			ModelName: "Basic",
			Fields: map[string]string{
				"Front": fields[0],
				"Back":  fields[1],
			},
			Options: map[string]interface{}{
				"allowDuplicate": false,
			},
		}

		// Parse tags if present
		if tags != "" {
			note.Tags = strings.Fields(tags)
		}

		// Extract media references from card content if syncMedia is enabled
		if syncMedia {
			note.Audio = extractMediaReferences(fields[0], fields[1], "sound")
			note.Picture = extractMediaReferences(fields[0], fields[1], "img")
			note.Video = extractMediaReferences(fields[0], fields[1], "video")
		}

		if _, err := client.AddNote(note); err != nil {
			// Skip duplicates
			if err.Error() != "AnkiConnect error: cannot create note because it is a duplicate" {
				return fmt.Errorf("failed to add card: %w", err)
			}
		}
	}

	return rows.Err()
}

// extractMediaReferences extracts media filenames from card content
func extractMediaReferences(front, back string, mediaType string) []ankiMedia {
	var media []ankiMedia

	// Simple extraction - in production, use proper HTML parsing
	switch mediaType {
	case "sound":
		// Look for [sound:filename] patterns
		if idx := strings.Index(front, "[sound:"); idx >= 0 {
			end := strings.Index(front[idx:], "]")
			if end > 0 {
				filename := front[idx+7 : idx+end]
				media = append(media, ankiMedia{
					Filename: filename,
					Fields:   []string{"Front"},
				})
			}
		}
		if idx := strings.Index(back, "[sound:"); idx >= 0 {
			end := strings.Index(back[idx:], "]")
			if end > 0 {
				filename := back[idx+7 : idx+end]
				media = append(media, ankiMedia{
					Filename: filename,
					Fields:   []string{"Back"},
				})
			}
		}
	case "img":
		// Look for <img src="filename"> patterns
		if idx := strings.Index(front, `<img src="`); idx >= 0 {
			end := strings.Index(front[idx+10:], `"`)
			if end > 0 {
				filename := front[idx+10 : idx+10+end]
				media = append(media, ankiMedia{
					Filename: filename,
					Fields:   []string{"Front"},
				})
			}
		}
		if idx := strings.Index(back, `<img src="`); idx >= 0 {
			end := strings.Index(back[idx+10:], `"`)
			if end > 0 {
				filename := back[idx+10 : idx+10+end]
				media = append(media, ankiMedia{
					Filename: filename,
					Fields:   []string{"Back"},
				})
			}
		}
	}

	return media
}

// SyncToAnki performs a more sophisticated sync with options
func (d *Deck) SyncToAnki(client *AnkiConnect, opts *SyncOptions) error {
	// Use default options if none provided
	syncOpts := opts
	if syncOpts == nil {
		syncOpts = &SyncOptions{
			UpdateExisting: true,
			DeleteMissing:  false,
			SyncMedia:      false,
		}
	}

	// Check connection
	if err := client.Ping(); err != nil {
		return fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}

	// Create deck if needed
	if err := client.CreateDeck(d.name); err != nil {
		if err.Error() != "AnkiConnect error: deck already exists" {
			return fmt.Errorf("failed to create deck: %w", err)
		}
	}

	// Find existing notes in the deck
	query := fmt.Sprintf("deck:\"%s\"", d.name)
	existingNotes, err := client.FindNotes(query)
	if err != nil {
		return fmt.Errorf("failed to find existing notes: %w", err)
	}

	// If UpdateExisting is true and there are existing notes, update them
	if syncOpts.UpdateExisting && len(existingNotes) > 0 {
		// Get detailed info about existing notes
		notesInfo, err := client.GetNotesInfo(existingNotes)
		if err != nil {
			return fmt.Errorf("failed to get notes info: %w", err)
		}

		// Create a map of existing notes by content for quick lookup
		existingMap := make(map[string]int64)
		for _, noteInfo := range notesInfo {
			if fields, ok := noteInfo["fields"].(map[string]interface{}); ok {
				var front, back string
				if f, ok := fields["Front"].(map[string]interface{}); ok {
					if v, ok := f["value"].(string); ok {
						front = v
					}
				}
				if b, ok := fields["Back"].(map[string]interface{}); ok {
					if v, ok := b["value"].(string); ok {
						back = v
					}
				}
				if noteID, ok := noteInfo["noteId"].(float64); ok {
					key := front + "|" + back
					existingMap[key] = int64(noteID)
				}
			}
		}

		// Update existing notes and add new ones
		return d.syncWithExisting(client, existingMap, syncOpts.SyncMedia)
	}

	// No existing notes, just push all cards
	return d.PushToAnkiWithMedia(client, syncOpts.SyncMedia)
}
