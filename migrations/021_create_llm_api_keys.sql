-- Create LLM API Keys table for project-scoped AI features
CREATE TABLE IF NOT EXISTS llm_api_keys (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('anthropic')),
    api_key TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Ensure only one active key per project per provider
    UNIQUE(project_id, provider)
);

-- Index for efficient project lookups
CREATE INDEX IF NOT EXISTS idx_llm_api_keys_project_id ON llm_api_keys(project_id);
CREATE INDEX IF NOT EXISTS idx_llm_api_keys_project_active ON llm_api_keys(project_id, is_active);

-- Trigger to update updated_at timestamp
DROP TRIGGER IF EXISTS update_llm_api_keys_updated_at;
CREATE TRIGGER update_llm_api_keys_updated_at
    AFTER UPDATE ON llm_api_keys
    FOR EACH ROW
BEGIN
    UPDATE llm_api_keys SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
