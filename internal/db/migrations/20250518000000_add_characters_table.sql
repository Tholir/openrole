-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    class TEXT NOT NULL,
    race TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    hp INTEGER NOT NULL DEFAULT 0,
    max_hp INTEGER NOT NULL DEFAULT 0,
    ac INTEGER NOT NULL DEFAULT 10,
    stats TEXT NOT NULL DEFAULT '{}',
    skills TEXT NOT NULL DEFAULT '[]',
    spells TEXT NOT NULL DEFAULT '[]',
    inventory TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_characters_name ON characters (name);
CREATE INDEX IF NOT EXISTS idx_characters_class ON characters (class);

CREATE TRIGGER IF NOT EXISTS update_characters_updated_at
AFTER UPDATE ON characters
BEGIN
UPDATE characters SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_characters_updated_at;
DROP TABLE IF EXISTS characters;

-- +goose StatementEnd