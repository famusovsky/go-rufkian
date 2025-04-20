package model

type Dictionary struct {
	UserID string `db:"user_id"`
	Hash   string `db:"hash"`
	Apkg   []byte `db:"apkg"`
}
