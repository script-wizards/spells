CREATE TABLE encounters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    name TEXT,
    description TEXT,
    is_active BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE initiative_order (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    encounter_id INTEGER NOT NULL,
    npc_id INTEGER,
    character_name TEXT, -- for PCs or custom combatants
    initiative INTEGER NOT NULL,
    hp_current INTEGER,
    hp_max INTEGER,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (encounter_id) REFERENCES encounters(id) ON DELETE CASCADE,
    FOREIGN KEY (npc_id) REFERENCES npcs(id) ON DELETE CASCADE
);

CREATE INDEX idx_initiative_order_encounter ON initiative_order(encounter_id);
CREATE INDEX idx_initiative_order_initiative ON initiative_order(encounter_id, initiative DESC, id);