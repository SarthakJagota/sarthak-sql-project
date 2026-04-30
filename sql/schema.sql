PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS blood_types (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS donors (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT NOT NULL,
	blood_type_id INTEGER NOT NULL,
	phone         TEXT,
	city          TEXT,
	created_at    TEXT NOT NULL,
	deleted_at    TEXT,
	FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
);

CREATE TABLE IF NOT EXISTS recipients (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT NOT NULL,
	blood_type_id INTEGER NOT NULL,
	phone         TEXT,
	hospital      TEXT,
	created_at    TEXT NOT NULL,
	deleted_at    TEXT,
	FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
);

CREATE TABLE IF NOT EXISTS donations (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	donor_id      INTEGER NOT NULL,
	units         INTEGER NOT NULL,
	donation_date TEXT NOT NULL,
	expiry_date   TEXT NOT NULL,
	deleted_at    TEXT,
	FOREIGN KEY(donor_id) REFERENCES donors(id)
);

CREATE TABLE IF NOT EXISTS inventory (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	blood_type_id INTEGER NOT NULL UNIQUE,
	units         INTEGER NOT NULL,
	deleted_at    TEXT,
	FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
);

CREATE TABLE IF NOT EXISTS requests (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	recipient_id INTEGER NOT NULL,
	units        INTEGER NOT NULL,
	status       TEXT NOT NULL,
	request_date TEXT NOT NULL,
	deleted_at   TEXT,
	FOREIGN KEY(recipient_id) REFERENCES recipients(id)
);
