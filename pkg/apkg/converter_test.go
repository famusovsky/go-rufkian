package apkg

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	notes := []SimpleNote{
		{Front: "test1", Back: "answer1"},
		{Front: "test2", Back: "answer2"},
	}
	anki2 := NewSimpleAnki(notes)

	result, err := Convert(anki2)

	require.NoError(t, err)
	require.NotNil(t, result)

	reader, err := zip.NewReader(bytes.NewReader(result), int64(len(result)))
	require.NoError(t, err)

	assert.Equal(t, 2, len(reader.File))

	var foundAnki2, foundMedia bool
	for _, file := range reader.File {
		if file.Name == "collection.anki2" {
			foundAnki2 = true
		} else if file.Name == "media" {
			foundMedia = true

			f, err := file.Open()
			require.NoError(t, err)

			content, err := io.ReadAll(f)
			require.NoError(t, err)

			assert.Equal(t, "{}", string(content))
			f.Close()
		}
	}

	assert.True(t, foundAnki2)
	assert.True(t, foundMedia)
}

func TestNewSimpleAnki(t *testing.T) {
	notes := []SimpleNote{
		{Front: "test1", Back: "answer1"},
		{Front: "test2", Back: "answer2"},
	}

	result := NewSimpleAnki(notes)

	assert.Equal(t, 2, len(result.Notes))
	assert.Equal(t, 2, len(result.Cards))
	assert.Equal(t, 0, len(result.Graves))
	assert.Equal(t, 0, len(result.Revlog))

	assert.Equal(t, 1, result.Col.ID)
	assert.Equal(t, 11, result.Col.Ver)

	for i, note := range result.Notes {
		assert.Equal(t, 1, note.Mid)
		assert.Contains(t, note.Flds, notes[i].Front)
		assert.Contains(t, note.Flds, notes[i].Back)
	}

	for i, card := range result.Cards {
		assert.Equal(t, result.Notes[i].ID, card.Nid)
		assert.Equal(t, 1, card.Did)
		assert.Equal(t, i, card.Due)
	}
}

func TestWrapIntoAPKG(t *testing.T) {
	anki2Data := []byte("test anki2 data")

	result, err := wrapIntoAPKG(anki2Data)

	require.NoError(t, err)
	require.NotNil(t, result)

	reader, err := zip.NewReader(bytes.NewReader(result), int64(len(result)))
	require.NoError(t, err)

	assert.Equal(t, 2, len(reader.File))

	var foundAnki2, foundMedia bool
	for _, file := range reader.File {
		if file.Name == "collection.anki2" {
			foundAnki2 = true

			f, err := file.Open()
			require.NoError(t, err)

			content, err := io.ReadAll(f)
			require.NoError(t, err)

			assert.Equal(t, anki2Data, content)
			f.Close()
		} else if file.Name == "media" {
			foundMedia = true

			f, err := file.Open()
			require.NoError(t, err)

			content, err := io.ReadAll(f)
			require.NoError(t, err)

			assert.Equal(t, "{}", string(content))
			f.Close()
		}
	}

	assert.True(t, foundAnki2)
	assert.True(t, foundMedia)
}

func TestWriteFile(t *testing.T) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	content := []byte("test content")
	fileName := "test.txt"

	err := writeFile(w, content, fileName)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	assert.Equal(t, 1, len(reader.File))
	assert.Equal(t, fileName, reader.File[0].Name)

	f, err := reader.File[0].Open()
	require.NoError(t, err)

	fileContent, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Equal(t, content, fileContent)
	f.Close()
}

func TestAnki2GetTable(t *testing.T) {
	notes := []SimpleNote{
		{Front: "test1", Back: "answer1"},
		{Front: "test2", Back: "answer2"},
	}

	anki2 := NewSimpleAnki(notes)

	cardsTable := anki2.GetTable("cards")
	assert.Equal(t, 2, len(cardsTable))

	colTable := anki2.GetTable("col")
	assert.Equal(t, 1, len(colTable))

	gravesTable := anki2.GetTable("graves")
	assert.Equal(t, 0, len(gravesTable))

	notesTable := anki2.GetTable("notes")
	assert.Equal(t, 2, len(notesTable))

	revlogTable := anki2.GetTable("revlog")
	assert.Equal(t, 0, len(revlogTable))

	unknownTable := anki2.GetTable("unknown")
	assert.Equal(t, 0, len(unknownTable))
}
