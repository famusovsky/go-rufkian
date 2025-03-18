package database

// user
const (
	addUserQuery        = `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING *;`
	updateUserQuery     = `UPDATE users SET email = $2, password = $3, key = $4 WHERE id = $1;`
	getUserQuery        = `SELECT * FROM users WHERE id = $1;`
	getUserByCredsQuery = `SELECT * FROM users WHERE email = $1;`
)

// dialog
const (
	getDialogQuery      = `SELECT * FROM dialogs WHERE id = $1;`
	getUserDialogsQuery = `SELECT * FROM dialogs WHERE user_id = $1;`
)

// dictionary
const (
	addUserWordQuery        = `INSERT INTO user_words (user_id, word) VALUES ($1, $2);`
	getUserWordsQuery       = `SELECT word FROM user_words WHERE user_id = $1;`
	checkUserWordQuery      = `SELECT COUNT(*) > 0 FROM user_words WHERE user_id = $1 AND word = $2;`
	deleteWordFromUserQuery = `DELETE FROM user_words WHERE user_id = $1 AND word = $2;`
)
