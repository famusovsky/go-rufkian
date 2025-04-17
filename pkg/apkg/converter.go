package apkg

import (
	"archive/zip"
	"bytes"

	"github.com/alicebob/bakelite"
)

func Convert(deck Deck) ([]byte, error) {
	db := bakelite.New()

	// TODO make with actual schema, it is just a test
	deckSlice := make([][]any, 0, len(deck))
	for _, card := range deck {
		if len(card) < 2 {
			continue
		}
		deckSlice = append(deckSlice, []any{card[0], card[1]})
	}

	db.AddSlice("cards", []string{"face", "back"}, deckSlice)

	buf := new(bytes.Buffer)
	if err := db.WriteTo(buf); err != nil {
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, err
	}

	return wrapIntoAPKG(buf.Bytes())
}

func wrapIntoAPKG(anki2 []byte) ([]byte, error) {
	res := new(bytes.Buffer)
	w := zip.NewWriter(res)

	f, err := w.Create("collection.anki2")
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(anki2); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}
