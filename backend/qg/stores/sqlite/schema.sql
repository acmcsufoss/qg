CREATE TABLE games (
	id TEXT PRIMARY KEY,
	typ TEXT NOT NULL,
	mod_password TEXT,
	data BLOB NOT NULL
);
