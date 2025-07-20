package anki

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const separator = "\u001F"

// Deck represents an Anki deck that can be exported as .apkg
type Deck struct {
	name       string
	db         *sql.DB
	media      []Media
	topDeckID  int64
	topModelID int64
}

// Media represents a media file to be included in the deck
type Media struct {
	Filename string
	Data     []byte
}

// CardOptions represents optional parameters for adding cards
type CardOptions struct {
	Tags []string
}

// TemplateOptions allows customization of card templates
type TemplateOptions struct {
	QuestionFormat string
	AnswerFormat   string
	CSS            string
}

// NewDeck creates a new Anki deck with the given name
func NewDeck(name string) (*Deck, error) {
	return NewDeckWithTemplate(name, nil)
}

// NewDeckWithTemplate creates a new Anki deck with custom template options
func NewDeckWithTemplate(name string, templateOpts *TemplateOptions) (*Deck, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	deck := &Deck{
		name:  name,
		db:    db,
		media: []Media{},
	}

	if err := deck.initializeDatabase(templateOpts); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return deck, nil
}

// AddCard adds a new card to the deck
func (d *Deck) AddCard(front, back string) error {
	return d.AddCardWithOptions(front, back, nil)
}

// AddCardWithOptions adds a new card with optional parameters
func (d *Deck) AddCardWithOptions(front, back string, opts *CardOptions) error {
	now := time.Now().UnixMilli()
	noteGUID := d.getNoteGUID(d.topDeckID, front, back)
	noteID := d.getNoteID(noteGUID, now)

	var tagsStr string
	if opts != nil && len(opts.Tags) > 0 {
		tags := make([]string, len(opts.Tags))
		for i, tag := range opts.Tags {
			tags[i] = strings.ReplaceAll(tag, " ", "_")
		}
		tagsStr = " " + strings.Join(tags, " ") + " "
	}

	// Insert or update note
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO notes 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		noteID,                           // id
		noteGUID,                         // guid
		d.topModelID,                     // mid
		d.getID("notes", "mod", now),     // mod
		-1,                               // usn
		tagsStr,                          // tags
		front+separator+back,             // flds
		front,                            // sfld
		d.checksum(front+separator+back), // csum
		0,                                // flags
		"",                               // data
	)
	if err != nil {
		return fmt.Errorf("failed to insert note: %w", err)
	}

	// Insert or update card
	_, err = d.db.Exec(`
		INSERT OR REPLACE INTO cards 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.getCardID(noteID, now),     // id
		noteID,                       // nid
		d.topDeckID,                  // did
		0,                            // ord
		d.getID("cards", "mod", now), // mod
		-1,                           // usn
		0,                            // type
		0,                            // queue
		179,                          // due
		0,                            // ivl
		0,                            // factor
		0,                            // reps
		0,                            // lapses
		0,                            // left
		0,                            // odue
		0,                            // odid
		0,                            // flags
		"",                           // data
	)
	if err != nil {
		return fmt.Errorf("failed to insert card: %w", err)
	}

	return nil
}

// AddMedia adds a media file to the deck
func (d *Deck) AddMedia(filename string, data []byte) {
	d.media = append(d.media, Media{
		Filename: filename,
		Data:     data,
	})
}

// Save exports the deck as an .apkg file
func (d *Deck) Save() ([]byte, error) {
	// Export database
	var dbData bytes.Buffer
	if err := d.exportDatabase(&dbData); err != nil {
		return nil, fmt.Errorf("failed to export database: %w", err)
	}

	// Create ZIP archive
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Add collection.anki2
	f, err := w.Create("collection.anki2")
	if err != nil {
		return nil, fmt.Errorf("failed to create collection.anki2: %w", err)
	}
	if _, err := f.Write(dbData.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to write collection.anki2: %w", err)
	}

	// Add media manifest
	mediaMap := make(map[string]string)
	for i, m := range d.media {
		mediaMap[strconv.Itoa(i)] = m.Filename
	}
	mediaJSON, err := json.Marshal(mediaMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal media map: %w", err)
	}

	f, err = w.Create("media")
	if err != nil {
		return nil, fmt.Errorf("failed to create media file: %w", err)
	}
	if _, err := f.Write(mediaJSON); err != nil {
		return nil, fmt.Errorf("failed to write media file: %w", err)
	}

	// Add media files
	for i, m := range d.media {
		f, err := w.Create(strconv.Itoa(i))
		if err != nil {
			return nil, fmt.Errorf("failed to create media file %d: %w", i, err)
		}
		if _, err := f.Write(m.Data); err != nil {
			return nil, fmt.Errorf("failed to write media file %d: %w", i, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Close closes the deck and releases resources
func (d *Deck) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Deck) initializeDatabase(templateOpts *TemplateOptions) error {
	template := createTemplate(templateOpts)
	if _, err := d.db.Exec(template); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	now := time.Now().UnixMilli()
	d.topDeckID = d.getID("cards", "did", now)
	d.topModelID = d.getID("notes", "mid", now)

	// Update deck name
	if err := d.updateDeckName(); err != nil {
		return fmt.Errorf("failed to update deck name: %w", err)
	}

	// Update model
	if err := d.updateModel(); err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	return nil
}

func (d *Deck) updateDeckName() error {
	var decksJSON string
	err := d.db.QueryRow("SELECT decks FROM col WHERE id = 1").Scan(&decksJSON)
	if err != nil {
		return err
	}

	var decks map[string]interface{}
	if err := json.Unmarshal([]byte(decksJSON), &decks); err != nil {
		return err
	}

	// Get the last deck and update it
	var lastKey string
	for k := range decks {
		lastKey = k
	}

	if lastKey != "" && lastKey != "1" {
		deck := decks[lastKey].(map[string]interface{})
		deck["name"] = d.name
		deck["id"] = float64(d.topDeckID)
		delete(decks, lastKey)
		decks[strconv.FormatInt(d.topDeckID, 10)] = deck
	}

	updatedJSON, err := json.Marshal(decks)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("UPDATE col SET decks = ? WHERE id = 1", string(updatedJSON))
	return err
}

func (d *Deck) updateModel() error {
	var modelsJSON string
	err := d.db.QueryRow("SELECT models FROM col WHERE id = 1").Scan(&modelsJSON)
	if err != nil {
		return err
	}

	var models map[string]interface{}
	if err := json.Unmarshal([]byte(modelsJSON), &models); err != nil {
		return err
	}

	// Get the last model and update it
	var lastKey string
	for k := range models {
		lastKey = k
	}

	if lastKey != "" {
		model := models[lastKey].(map[string]interface{})
		model["name"] = d.name
		model["did"] = float64(d.topDeckID)
		model["id"] = float64(d.topModelID)
		delete(models, lastKey)
		models[strconv.FormatInt(d.topModelID, 10)] = model
	}

	updatedJSON, err := json.Marshal(models)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("UPDATE col SET models = ? WHERE id = 1", string(updatedJSON))
	return err
}

func (d *Deck) getID(table, col string, ts int64) int64 {
	var maxID sql.NullInt64
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s >= ? ORDER BY %s DESC LIMIT 1", col, table, col, col)
	err := d.db.QueryRow(query, ts).Scan(&maxID)
	if err != nil || !maxID.Valid {
		return ts
	}
	return maxID.Int64 + 1
}

func (d *Deck) getNoteID(guid string, ts int64) int64 {
	var id sql.NullInt64
	err := d.db.QueryRow("SELECT id FROM notes WHERE guid = ? ORDER BY id DESC LIMIT 1", guid).Scan(&id)
	if err != nil || !id.Valid {
		return d.getID("notes", "id", ts)
	}
	return id.Int64
}

func (d *Deck) getNoteGUID(deckID int64, front, back string) string {
	data := fmt.Sprintf("%d%s%s", deckID, front, back)
	return fmt.Sprintf("%x", sha1.Sum([]byte(data)))
}

func (d *Deck) getCardID(noteID, ts int64) int64 {
	var id sql.NullInt64
	err := d.db.QueryRow("SELECT id FROM cards WHERE nid = ? ORDER BY id DESC LIMIT 1", noteID).Scan(&id)
	if err != nil || !id.Valid {
		return d.getID("cards", "id", ts)
	}
	return id.Int64
}

func (d *Deck) checksum(str string) int64 {
	hash := sha1.Sum([]byte(str))
	// Take first 8 characters of hex and convert to int64
	hexStr := fmt.Sprintf("%x", hash)[:8]
	val, _ := strconv.ParseInt(hexStr, 16, 64)
	return val
}
