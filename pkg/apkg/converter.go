package apkg

import (
	"archive/zip"
	"bytes"

	"github.com/alicebob/bakelite"
)

func Convert(anki2 Anki2) ([]byte, error) {
	db := bakelite.New()
	defer db.Close()

	for table, columns := range tables {
		if err := db.AddSlice(table, columns, anki2.GetTable(table)); err != nil {
			return nil, err
		}
	}

	buf := new(bytes.Buffer)
	if err := db.WriteTo(buf); err != nil {
		return nil, err
	}

	return wrapIntoAPKG(buf.Bytes())
}

func wrapIntoAPKG(anki2 []byte) ([]byte, error) {
	res := new(bytes.Buffer)
	w := zip.NewWriter(res)

	writeFile(w, anki2, "collection.anki2")
	writeFile(w, []byte("{}"), "media") // TODO add media support

	if err := w.Close(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func writeFile(w *zip.Writer, contents []byte, name string) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}

	if _, err := f.Write(contents); err != nil {
		return err
	}

	return nil
}
