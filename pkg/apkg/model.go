package apkg

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"
)

type SimpleNote struct {
	Front, Back string
}

func NewSimpleAnki(notes []SimpleNote) Anki2 {
	res := Anki2{
		Graves: make([]Grave, 0),
		Revlog: make([]Revlog, 0),
	}

	now := time.Now()
	unix := int(now.Unix())
	unixMilli := int(now.UnixMilli())

	res.Col = Col{
		ID:     1,
		Crt:    unix,
		Mod:    unixMilli,
		Scm:    unixMilli,
		Ver:    11,
		Conf:   defaultColConf,
		Models: defaultColModels,
		Decks:  defaultColDecks,
		Dconf:  defaultColDConf,
		Tags:   "{}",
	}

	res.Notes = make([]Note, 0, len(notes))
	res.Cards = make([]Card, 0, len(notes))
	for i, note := range notes {
		res.Notes = append(res.Notes, Note{
			ID:   unixMilli + i,
			Guid: strconv.Itoa(rand.Int()),
			Mid:  1,
			Mod:  unix,
			Usn:  unixMilli,
			Flds: fmt.Sprintf("%s\x1f%s", note.Front, note.Back),
			Csum: unixMilli,
		})

		res.Cards = append(res.Cards, Card{
			ID:  unixMilli + i,
			Nid: res.Notes[i].ID,
			Mod: unix,
			Did: 1,
			Usn: unixMilli,
			Due: i,
		})
	}

	return res
}

type Anki2 struct {
	Cards  []Card
	Col    Col
	Graves []Grave
	Notes  []Note
	Revlog []Revlog
}

func (a2 *Anki2) GetTable(table string) [][]any {
	var rows [][]any

	switch table {
	case "cards":
		rows = make([][]any, 0, len(a2.Cards))
		for _, card := range a2.Cards {
			rows = append(rows, card.toRow())
		}
	case "col":
		rows = [][]any{a2.Col.toRow()}
	case "graves":
		rows = make([][]any, 0, len(a2.Graves))
		for _, grave := range a2.Graves {
			rows = append(rows, grave.toRow())
		}
	case "notes":
		rows = make([][]any, 0, len(a2.Notes))
		for _, note := range a2.Notes {
			rows = append(rows, note.toRow())
		}
	case "revlog":
		rows = make([][]any, 0, len(a2.Revlog))
		for _, revlog := range a2.Revlog {
			rows = append(rows, revlog.toRow())
		}
	}

	return rows
}

// Card represents a row in the cards table
type Card struct {
	ID     int    // the epoch milliseconds of when the card was created
	Nid    int    // notes.id
	Did    int    // deck id (available in col table)
	Ord    int    // ordinal : identifies which of the card templates or cloze deletions it corresponds to
	Mod    int    // modification time as epoch seconds
	Usn    int    // update sequence number : used to figure out diffs when syncing.
	Type   int    // 0=new, 1=learning, 2=review, 3=relearning
	Queue  int    // -3=user buried(In scheduler 2), -2=sched buried (In scheduler 2), -2=buried(In scheduler 1), -1=suspended, 0=new, 1=learning, 2=review (as for type), 3=in learning, next rev in at least a day after the previous review, 4=preview
	Due    int    // Due is used differently for different card types: new: the order in which cards are to be studied; starts from 1. learning/relearning: epoch timestamp in seconds review: days since the collection's creation time
	Ivl    int    // interval (used in SRS algorithm). Negative = seconds, positive = days
	Factor int    // The ease factor of the card in permille (parts per thousand). If the ease factor is 2500, the card’s interval will be multiplied by 2.5 the next time you press Good.
	Reps   int    // number of reviews
	Lapses int    // the number of times the card went from a "was answered correctly" to "was answered incorrectly" state
	Left   int    // of the form a*1000+b, with: a the number of reps left today b the number of reps left till graduation
	Odue   int    // original due: In filtered decks, it's the original due date that the card had before moving to filtered.
	Odid   int    // original did: only used when the card is currently in filtered deck
	Flags  int    // an integer. This integer mod 8 represents a "flag", which can be see in browser and while reviewing a note. Red 1, Orange 2, Green 3, Blue 4, no flag: 0. This integer divided by 8 represents currently nothing
	Data   string // currently unused
}

func (c *Card) toRow() []any {
	return []any{
		c.ID, c.Nid, c.Did, c.Ord, c.Mod, c.Usn, c.Type, c.Queue, c.Due, c.Ivl, c.Factor, c.Reps, c.Lapses, c.Left, c.Odue, c.Odid, c.Flags, c.Data,
	}
}

// Col represents a row in the col table
type Col struct {
	ID     int    // arbitrary number since there is only one row
	Crt    int    // timestamp of the creation date in second. It's correct up to the day. For V1 scheduler, the hour corresponds to starting a new day. By default, new day is 4.
	Mod    int    // last modified in milliseconds
	Scm    int    // schema mod time: time when "schema" was modified.
	Ver    int    // version
	Dty    int    // dirty: unused, set to 0
	Usn    int    // update sequence number: used for finding diffs when syncing.
	Ls     int    // "last sync time"
	Conf   string // json object containing configuration options that are synced. Described below in "configuration JSONObjects"
	Models string // json object of json object(s) representing the models (aka Note types)
	Decks  string // json object of json object(s) representing the deck(s)
	Dconf  string // json object of json object(s) representing the options group(s) for decks
	Tags   string // a cache of tags used in the collection (This list is displayed in the browser. Potentially at other place)
}

func (c *Col) toRow() []any {
	return []any{
		c.ID, c.Crt, c.Mod, c.Scm, c.Ver, c.Dty, c.Usn, c.Ls, c.Conf, c.Models, c.Decks, c.Dconf, c.Tags,
	}
}

// Grave represents a row in the graves table
type Grave struct {
	Usn  int // usn should be set to -1
	Oid  int // oid is the original id
	Type int // type: 0 for a card, 1 for a note and 2 for a deck
}

func (g *Grave) toRow() []any {
	return []any{
		g.Usn, g.Oid, g.Type,
	}
}

// Note represents a row in the notes table
type Note struct {
	ID    int    // epoch milliseconds of when the note was created
	Guid  string // globally unique id, almost certainly used for syncing
	Mid   int    // model id
	Mod   int    // modification timestamp, epoch seconds
	Usn   int    // update sequence number: for finding diffs when syncing.
	Tags  string // space-separated string of tags.
	Flds  string // the values of the fields in this note. separated by 0x1f (31) character.
	Sfld  int    // sort field: used for quick sorting and duplicate check. The sort field is an integer so that when users are sorting on a field that contains only numbers, they are sorted in numeric instead of lexical order. Text is stored in this integer field.
	Csum  int    // field checksum used for duplicate check.
	Flags int    // unused
	Data  string // unused
}

func (n *Note) toRow() []any {
	return []any{
		n.ID, n.Guid, n.Mid, n.Mod, n.Usn, n.Tags, n.Flds, n.Sfld, n.Csum, n.Flags, n.Data,
	}
}

// Revlog represents a row in the revlog table
type Revlog struct {
	ID      int // epoch-milliseconds timestamp of when you did the review
	Cid     int // cards.id
	Usn     int // update sequence number: for finding diffs when syncing.
	Ease    int // which button you pushed to score your recall.
	Ivl     int // interval (i.e. as in the card table)
	LastIvl int // last interval (i.e. the last value of ivl. Note that this value is not necessarily equal to the actual interval between this review and the preceding review)
	Factor  int // factor
	Time    int // how many milliseconds your review took, up to 60000 (60s)
	Type    int // 0=learn, 1=review, 2=relearn, 3=filtered, 4=manual, 5=rescheduled
}

func (r *Revlog) toRow() []any {
	return []any{
		r.ID, r.Cid, r.Usn, r.Ease, r.Ivl, r.LastIvl, r.Factor, r.Time, r.Type,
	}
}
