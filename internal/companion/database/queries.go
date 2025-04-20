package database

// user
const (
	addUserQuery        = `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id;`
	updateUserQuery     = `UPDATE users SET email = $2, password = $3, key = $4, time_goal_m = $5 WHERE id = $1;`
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

	addWordQuery = `INSERT INTO words (word, info, translation) VALUES ($1, $2, $3);`
	getWordQuery = `SELECT info, translation FROM words WHERE word = $1;`

	getDictionaryQuery     = `SELECT * FROM dictionaries WHERE user_id = $1;`
	getDictionaryHashQuery = `SELECT hash FROM dictionaries WHERE user_id = $1;`
	updateDictionaryQuery  = `UPDATE dictionaries SET hash = $2, apkg = $3 WHERE user_id = $1;`
)
