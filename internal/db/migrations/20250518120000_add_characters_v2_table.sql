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
    str_score INTEGER NOT NULL DEFAULT 10,
    dex_score INTEGER NOT NULL DEFAULT 10,
    con_score INTEGER NOT NULL DEFAULT 10,
    int_score INTEGER NOT NULL DEFAULT 10,
    wis_score INTEGER NOT NULL DEFAULT 10,
    cha_score INTEGER NOT NULL DEFAULT 10,
    inventory_json TEXT NOT NULL DEFAULT '[]',
    skills_json TEXT NOT NULL DEFAULT '[]',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_characters_name ON characters (name);
CREATE INDEX IF NOT EXISTS idx_characters_class ON characters (class);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS characters;

-- +goose StatementEnd
