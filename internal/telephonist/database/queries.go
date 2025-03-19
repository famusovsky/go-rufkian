package database

const storeDialogQuery = "INSERT INTO dialogs (user_id, start_time, duration_s, messages) VALUES ($1, $2, $3, $4) RETURNING id;"
