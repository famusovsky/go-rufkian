package database

const storeDialogQuery = "INSERT INTO dialogs (user_id, start_time, messages) VALUES ($1, $2, $3) RETURNING id;"
