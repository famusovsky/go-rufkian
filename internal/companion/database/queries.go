package database

const (
	addUserQuery        = `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING *;`
	getUserQuery        = `SELECT * FROM users WHERE id = $1;`
	getUserByCredsQuery = `SELECT * FROM users WHERE email = $1;`
)
