package db

var (
	dbMigrations = []string{
		`
	CREATE TABLE metadata (
		key		text NOT NULL,
		value	text
	);
	CREATE UNIQUE INDEX idx_key ON metadata(key);
	INSERT INTO metadata (key, value) VALUES ('last_height', '0');
	INSERT INTO metadata (key, value) VALUES ('db_version', '1');
	CREATE TABLE accounts (
		id				INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		address_index	INTEGER NOT NULL,
		address			TEXT NOT NULL,
		user_name		TEXT,
		user_address	TEXT,
		amount			INTEGER NOT NULL DEFAULT 0,
		entries			INTEGER NOT NULL DEFAULT 0,
		active			INTEGER NOT NULL DEFAULT 1,
		ref_id			INTEGER NOT NULL DEFAULT 0
	);
	CREATE UNIQUE INDEX idx_addr_idx ON accounts(user_name);
	CREATE INDEX idx_user_address ON accounts(user_address);
	CREATE INDEX idx_active ON accounts(active);
	CREATE INDEX idx_ref_id ON accounts(ref_id);
	CREATE INDEX idx_entries ON accounts(entries);
	INSERT INTO metadata (key, value) VALUES ('entry_id', '0');
	INSERT INTO metadata (key, value)
		VALUES ('sign_key', '90a7e39da756fdb53c55c4e00ff05a70db9083b9f8cfca7354582f756b9d9edf');
	CREATE TABLE entries (
		id				INTEGER NOT NULL PRIMARY KEY,
		account_id		INTEGER NOT NULL,
		hash			TEXT NOT NULL
	);
	CREATE INDEX idx_acct_id ON entries(account_id);
	CREATE INDEX idx_hash ON entries(hash);
	CREATE TABLE winners (
		date			TEXT NOT NULL PRIMARY KEY,
		info			TEXT NOT NULL,
		transfer_body 	TEXT
	);
	CREATE TABLE transactions (
		id				TEXT NOT NULL PRIMARY KEY
	);`,
	}
)
