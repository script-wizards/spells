CREATE TABLE npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    location TEXT,
    status TEXT DEFAULT 'neutral', -- ally/neutral/hostile
    motivation TEXT,
    secrets TEXT,
    tags TEXT, -- JSON array
    last_mentioned TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);