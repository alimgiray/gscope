-- Migration 020: Remove email column from users table
-- Date: 2025-08-03
-- Description: Remove email field from users table as we're transitioning to username-based authentication

-- Remove the email column from users table
ALTER TABLE users DROP COLUMN email; 