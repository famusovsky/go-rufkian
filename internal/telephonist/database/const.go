package database

const storeDialogQuery = "INSERT INTO dialogs (key, messages) VALUES ($1, $2) RETURNING id;"
