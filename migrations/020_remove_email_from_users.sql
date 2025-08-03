-- Migration 020: Remove email column from users table
-- Date: 2025-08-03
-- Description: Remove email field from users table as we're transitioning to username-based authentication

-- SQLite doesn't support dropping UNIQUE columns directly, so we need to recreate the table
-- First, create a temporary table with the new schema
CREATE TABLE users_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    profile_picture TEXT,
    access_token TEXT,
    github_access_token TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table to new table (excluding email column)
INSERT INTO users_new (id, name, username, profile_picture, access_token, github_access_token, created_at)
SELECT id, name, username, profile_picture, access_token, github_access_token, created_at
FROM users;

-- Drop the old table
DROP TABLE users;

-- Rename the new table to the original name
ALTER TABLE users_new RENAME TO users; 