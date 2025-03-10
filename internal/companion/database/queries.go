package database

// user
const (
	addUserQuery        = `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING *;`
	getUserQuery        = `SELECT * FROM users WHERE id = $1;`
	getUserByCredsQuery = `SELECT * FROM users WHERE email = $1;`
)

// dialog
const (
	getDialogQuery      = `SELECT * FROM dialogs WHERE id = $1;`
	getUserDialogsQuery = `SELECT * FROM dialogs WHERE user_id = $1;`
)
