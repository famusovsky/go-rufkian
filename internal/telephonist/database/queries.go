package database

const storeDialogQuery = "INSERT INTO dialogs (user_id, messages) VALUES ($1, $2) RETURNING id;"
