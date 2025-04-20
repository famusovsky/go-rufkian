CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    key TEXT,
    time_goal_m INTEGER NOT NULL DEFAULT 0,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dialogs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    start_time TIMESTAMP NOT NULL,
    duration_s INTEGER NOT NULL,
    messages TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_words (
    user_id INTEGER NOT NULL,
    word TEXT NOT NULL,
    UNIQUE (user_id, word),
    FOREIGN KEY (user_id) REFERENCES users(id)
)

CREATE TABLE IF NOT EXISTS words (
    word TEXT PRIMARY KEY,
    info TEXT NOT NULL,
    translation TEXT NOT NULL
)

CREATE TABLE IF NOT EXISTS dictionaries (
    user_id INTEGER PRIMARY KEY,
    hash TEXT NOT NULL,
    apkg BYTEA NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
)