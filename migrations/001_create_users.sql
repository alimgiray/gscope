CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    profile_picture TEXT,
    access_token TEXT,
    github_access_token TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
); 