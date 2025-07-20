package anki

import (
	"encoding/json"
	"fmt"
)

func createTemplate(opts *TemplateOptions) string {
	if opts == nil {
		opts = &TemplateOptions{}
	}

	// Set defaults
	if opts.QuestionFormat == "" {
		opts.QuestionFormat = "{{Front}}"
	}
	if opts.AnswerFormat == "" {
		opts.AnswerFormat = "{{FrontSide}}\n\n<hr id=\"answer\">\n\n{{Back}}"
	}
	if opts.CSS == "" {
		opts.CSS = `.card {
 font-family: arial;
 font-size: 20px;
 text-align: center;
 color: black;
background-color: white;
}`
	}

	conf := map[string]interface{}{
		"nextPos":       1,
		"estTimes":      true,
		"activeDecks":   []int{1},
		"sortType":      "noteFld",
		"timeLim":       0,
		"sortBackwards": false,
		"addToCur":      true,
		"curDeck":       1,
		"newBury":       true,
		"newSpread":     0,
		"dueCounts":     true,
		"curModel":      "1435645724216",
		"collapseTime":  1200,
	}

	models := map[string]interface{}{
		"1388596687391": map[string]interface{}{
			"vers": []interface{}{},
			"name": "Basic-f15d2",
			"tags": []string{"Tag"},
			"did":  1435588830424,
			"usn":  -1,
			"req":  [][]interface{}{{0, "all", []int{0}}},
			"flds": []map[string]interface{}{
				{
					"name":   "Front",
					"media":  []interface{}{},
					"sticky": false,
					"rtl":    false,
					"ord":    0,
					"font":   "Arial",
					"size":   20,
				},
				{
					"name":   "Back",
					"media":  []interface{}{},
					"sticky": false,
					"rtl":    false,
					"ord":    1,
					"font":   "Arial",
					"size":   20,
				},
			},
			"sortf":    0,
			"latexPre": "\\documentclass[12pt]{article}\n\\special{papersize=3in,5in}\n\\usepackage[utf8]{inputenc}\n\\usepackage{amssymb,amsmath}\n\\pagestyle{empty}\n\\setlength{\\parindent}{0in}\n\\begin{document}\n",
			"tmpls": []map[string]interface{}{
				{
					"name":  "Card 1",
					"qfmt":  opts.QuestionFormat,
					"did":   nil,
					"bafmt": "",
					"afmt":  opts.AnswerFormat,
					"ord":   0,
					"bqfmt": "",
				},
			},
			"latexPost": "\\end{document}",
			"type":      0,
			"id":        1388596687391,
			"css":       opts.CSS,
			"mod":       1435645658,
		},
	}

	decks := map[string]interface{}{
		"1": map[string]interface{}{
			"desc":      "",
			"name":      "Default",
			"extendRev": 50,
			"usn":       0,
			"collapsed": false,
			"newToday":  []int{0, 0},
			"timeToday": []int{0, 0},
			"dyn":       0,
			"extendNew": 10,
			"conf":      1,
			"revToday":  []int{0, 0},
			"lrnToday":  []int{0, 0},
			"id":        1,
			"mod":       1435645724,
		},
		"1435588830424": map[string]interface{}{
			"desc":      "",
			"name":      "Template",
			"extendRev": 50,
			"usn":       -1,
			"collapsed": false,
			"newToday":  []int{545, 0},
			"timeToday": []int{545, 0},
			"dyn":       0,
			"extendNew": 10,
			"conf":      1,
			"revToday":  []int{545, 0},
			"lrnToday":  []int{545, 0},
			"id":        1435588830424,
			"mod":       1435588830,
		},
	}

	dconf := map[string]interface{}{
		"1": map[string]interface{}{
			"name":    "Default",
			"replayq": true,
			"lapse": map[string]interface{}{
				"leechFails":  8,
				"minInt":      1,
				"delays":      []int{10},
				"leechAction": 0,
				"mult":        0,
			},
			"rev": map[string]interface{}{
				"perDay":   100,
				"fuzz":     0.05,
				"ivlFct":   1,
				"maxIvl":   36500,
				"ease4":    1.3,
				"bury":     true,
				"minSpace": 1,
			},
			"timer":    0,
			"maxTaken": 60,
			"usn":      0,
			"new": map[string]interface{}{
				"perDay":        20,
				"delays":        []int{1, 10},
				"separate":      true,
				"ints":          []int{1, 4, 7},
				"initialFactor": 2500,
				"bury":          true,
				"order":         1,
			},
			"mod":      0,
			"id":       1,
			"autoplay": true,
		},
	}

	confJSON, _ := json.Marshal(conf)
	modelsJSON, _ := json.Marshal(models)
	decksJSON, _ := json.Marshal(decks)
	dconfJSON, _ := json.Marshal(dconf)

	return fmt.Sprintf(`
    PRAGMA foreign_keys=OFF;
    BEGIN TRANSACTION;
    CREATE TABLE col (
        id              integer primary key,
        crt             integer not null,
        mod             integer not null,
        scm             integer not null,
        ver             integer not null,
        dty             integer not null,
        usn             integer not null,
        ls              integer not null,
        conf            text not null,
        models          text not null,
        decks           text not null,
        dconf           text not null,
        tags            text not null
    );
    INSERT INTO "col" VALUES(
      1,
      1388548800,
      1435645724219,
      1435645724215,
      11,
      0,
      0,
      0,
      '%s',
      '%s',
      '%s',
      '%s',
      '{}'
    );
    CREATE TABLE notes (
        id              integer primary key,   /* 0 */
        guid            text not null,         /* 1 */
        mid             integer not null,      /* 2 */
        mod             integer not null,      /* 3 */
        usn             integer not null,      /* 4 */
        tags            text not null,         /* 5 */
        flds            text not null,         /* 6 */
        sfld            integer not null,      /* 7 */
        csum            integer not null,      /* 8 */
        flags           integer not null,      /* 9 */
        data            text not null          /* 10 */
    );
    CREATE TABLE cards (
        id              integer primary key,   /* 0 */
        nid             integer not null,      /* 1 */
        did             integer not null,      /* 2 */
        ord             integer not null,      /* 3 */
        mod             integer not null,      /* 4 */
        usn             integer not null,      /* 5 */
        type            integer not null,      /* 6 */
        queue           integer not null,      /* 7 */
        due             integer not null,      /* 8 */
        ivl             integer not null,      /* 9 */
        factor          integer not null,      /* 10 */
        reps            integer not null,      /* 11 */
        lapses          integer not null,      /* 12 */
        left            integer not null,      /* 13 */
        odue            integer not null,      /* 14 */
        odid            integer not null,      /* 15 */
        flags           integer not null,      /* 16 */
        data            text not null          /* 17 */
    );
    CREATE TABLE revlog (
        id              integer primary key,
        cid             integer not null,
        usn             integer not null,
        ease            integer not null,
        ivl             integer not null,
        lastIvl         integer not null,
        factor          integer not null,
        time            integer not null,
        type            integer not null
    );
    CREATE TABLE graves (
        usn             integer not null,
        oid             integer not null,
        type            integer not null
    );
    ANALYZE sqlite_master;
    INSERT INTO "sqlite_stat1" VALUES('col',NULL,'1');
    CREATE INDEX ix_notes_usn on notes (usn);
    CREATE INDEX ix_cards_usn on cards (usn);
    CREATE INDEX ix_revlog_usn on revlog (usn);
    CREATE INDEX ix_cards_nid on cards (nid);
    CREATE INDEX ix_cards_sched on cards (did, queue, due);
    CREATE INDEX ix_revlog_cid on revlog (cid);
    CREATE INDEX ix_notes_csum on notes (csum);
    COMMIT;
  `, string(confJSON), string(modelsJSON), string(decksJSON), string(dconfJSON))
}
