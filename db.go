package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func initDB(db *sql.DB) error {
	schemaSQL, err := assets.ReadFile("sql/schema.sql")
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return err
	}
	if err := migrateTo3NF(db); err != nil {
		return err
	}
	for _, table := range []string{"donors", "recipients", "donations", "inventory", "requests"} {
		if err := ensureColumn(db, table, "deleted_at", "TEXT"); err != nil {
			return err
		}
	}
	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_blood_type_id ON inventory(blood_type_id)")
	return err
}

func ensureColumn(db *sql.DB, table, column, colType string) error {
	exists, err := tableHasColumn(db, table, column)
	if err != nil || exists {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colType))
	return err
}

func tableHasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func normalizeBloodType(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func getOrCreateBloodTypeID(db *sql.DB, bloodType string) (int, error) {
	bloodType = normalizeBloodType(bloodType)
	if bloodType == "" {
		return 0, fmt.Errorf("blood type required")
	}
	var id int
	err := db.QueryRow("SELECT id FROM blood_types WHERE type = ?", bloodType).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := db.Exec("INSERT INTO blood_types (type) VALUES (?)", bloodType)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func getDonorBloodTypeID(db *sql.DB, donorID int) (int, error) {
	var id int
	if err := db.QueryRow("SELECT blood_type_id FROM donors WHERE id = ? AND deleted_at IS NULL", donorID).Scan(&id); err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("missing blood type")
	}
	return id, nil
}

func getRecipientBloodTypeID(db *sql.DB, recipientID int) (int, error) {
	var id int
	if err := db.QueryRow("SELECT blood_type_id FROM recipients WHERE id = ? AND deleted_at IS NULL", recipientID).Scan(&id); err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("missing blood type")
	}
	return id, nil
}

func getRequestBloodTypeID(db *sql.DB, requestID int) (int, error) {
	var id int
	err := db.QueryRow(`
		SELECT recipients.blood_type_id
		FROM requests r
		JOIN recipients ON recipients.id = r.recipient_id
		WHERE r.id = ? AND r.deleted_at IS NULL
	`, requestID).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("missing blood type")
	}
	return id, nil
}

func upsertInventoryByTypeID(db *sql.DB, bloodTypeID, units int) error {
	res, err := db.Exec("UPDATE inventory SET units = units + ?, deleted_at = NULL WHERE blood_type_id = ?", units, bloodTypeID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		_, err = db.Exec("INSERT INTO inventory (blood_type_id, units) VALUES (?, ?)", bloodTypeID, units)
	}
	return err
}

func consumeInventoryByTypeID(db *sql.DB, bloodTypeID, units int) (bool, error) {
	var current int
	err := db.QueryRow("SELECT units FROM inventory WHERE blood_type_id = ? AND deleted_at IS NULL", bloodTypeID).Scan(&current)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if current < units {
		return false, nil
	}
	_, err = db.Exec("UPDATE inventory SET units = units - ? WHERE blood_type_id = ?", units, bloodTypeID)
	return err == nil, err
}

func loadPageData(db *sql.DB, msg string) (PageData, error) {
	data := PageData{Message: msg}
	var err error
	if data.Donors, err = loadDonors(db); err != nil {
		return data, err
	}
	if data.Recipients, err = loadRecipients(db); err != nil {
		return data, err
	}
	if data.Donations, err = loadDonations(db); err != nil {
		return data, err
	}
	if data.Inventory, err = loadInventory(db); err != nil {
		return data, err
	}
	if data.Requests, err = loadRequests(db); err != nil {
		return data, err
	}
	return data, nil
}

func loadDonors(db *sql.DB) ([]Donor, error) {
	rows, err := db.Query(`
		SELECT d.id, d.name, bt.type, d.phone, d.city, d.created_at
		FROM donors d
		JOIN blood_types bt ON bt.id = d.blood_type_id
		WHERE d.deleted_at IS NULL
		ORDER BY d.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var donors []Donor
	for rows.Next() {
		var d Donor
		if err := rows.Scan(&d.ID, &d.Name, &d.BloodType, &d.Phone, &d.City, &d.CreatedAt); err != nil {
			return nil, err
		}
		donors = append(donors, d)
	}
	return donors, rows.Err()
}

func loadRecipients(db *sql.DB) ([]Recipient, error) {
	rows, err := db.Query(`
		SELECT r.id, r.name, bt.type, r.phone, r.hospital, r.created_at
		FROM recipients r
		JOIN blood_types bt ON bt.id = r.blood_type_id
		WHERE r.deleted_at IS NULL
		ORDER BY r.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recipients []Recipient
	for rows.Next() {
		var r Recipient
		if err := rows.Scan(&r.ID, &r.Name, &r.BloodType, &r.Phone, &r.Hospital, &r.CreatedAt); err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}
	return recipients, rows.Err()
}

func loadDonations(db *sql.DB) ([]Donation, error) {
	rows, err := db.Query(`
		SELECT d.id, d.donor_id, donors.name, bt.type, d.units, d.donation_date, d.expiry_date
		FROM donations d
		JOIN donors ON donors.id = d.donor_id
		JOIN blood_types bt ON bt.id = donors.blood_type_id
		WHERE d.deleted_at IS NULL
		ORDER BY d.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var donations []Donation
	for rows.Next() {
		var d Donation
		if err := rows.Scan(&d.ID, &d.DonorID, &d.DonorName, &d.BloodType, &d.Units, &d.DonationDate, &d.ExpiryDate); err != nil {
			return nil, err
		}
		donations = append(donations, d)
	}
	return donations, rows.Err()
}

func loadInventory(db *sql.DB) ([]Inventory, error) {
	rows, err := db.Query(`
		SELECT bt.type, i.units
		FROM inventory i
		JOIN blood_types bt ON bt.id = i.blood_type_id
		WHERE i.deleted_at IS NULL
		ORDER BY bt.type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var inv []Inventory
	for rows.Next() {
		var i Inventory
		if err := rows.Scan(&i.BloodType, &i.Units); err != nil {
			return nil, err
		}
		inv = append(inv, i)
	}
	return inv, rows.Err()
}

func loadRequests(db *sql.DB) ([]Request, error) {
	rows, err := db.Query(`
		SELECT r.id, r.recipient_id, recipients.name, bt.type, r.units, r.status, r.request_date
		FROM requests r
		JOIN recipients ON recipients.id = r.recipient_id
		JOIN blood_types bt ON bt.id = recipients.blood_type_id
		WHERE r.deleted_at IS NULL
		ORDER BY r.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var requests []Request
	for rows.Next() {
		var r Request
		if err := rows.Scan(&r.ID, &r.RecipientID, &r.Recipient, &r.BloodType, &r.Units, &r.Status, &r.RequestDate); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func migrateTo3NF(db *sql.DB) error {
	hasBloodType, err := tableHasColumn(db, "donors", "blood_type")
	if err != nil || !hasBloodType {
		return err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS blood_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL UNIQUE
		)`,
		`INSERT OR IGNORE INTO blood_types (type)
			SELECT DISTINCT UPPER(TRIM(blood_type)) FROM donors WHERE blood_type IS NOT NULL AND TRIM(blood_type) <> ''`,
		`INSERT OR IGNORE INTO blood_types (type)
			SELECT DISTINCT UPPER(TRIM(blood_type)) FROM recipients WHERE blood_type IS NOT NULL AND TRIM(blood_type) <> ''`,
		`INSERT OR IGNORE INTO blood_types (type)
			SELECT DISTINCT UPPER(TRIM(blood_type)) FROM inventory WHERE blood_type IS NOT NULL AND TRIM(blood_type) <> ''`,
		`INSERT OR IGNORE INTO blood_types (type) VALUES ('UNKNOWN')`,
		`CREATE TABLE donors_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			blood_type_id INTEGER NOT NULL,
			phone TEXT,
			city TEXT,
			created_at TEXT NOT NULL,
			deleted_at TEXT,
			FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
		)`,
		`CREATE TABLE recipients_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			blood_type_id INTEGER NOT NULL,
			phone TEXT,
			hospital TEXT,
			created_at TEXT NOT NULL,
			deleted_at TEXT,
			FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
		)`,
		`CREATE TABLE donations_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			donor_id INTEGER NOT NULL,
			units INTEGER NOT NULL,
			donation_date TEXT NOT NULL,
			expiry_date TEXT NOT NULL,
			deleted_at TEXT,
			FOREIGN KEY(donor_id) REFERENCES donors(id)
		)`,
		`CREATE TABLE requests_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			recipient_id INTEGER NOT NULL,
			units INTEGER NOT NULL,
			status TEXT NOT NULL,
			request_date TEXT NOT NULL,
			deleted_at TEXT,
			FOREIGN KEY(recipient_id) REFERENCES recipients(id)
		)`,
		`CREATE TABLE inventory_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			blood_type_id INTEGER NOT NULL UNIQUE,
			units INTEGER NOT NULL,
			deleted_at TEXT,
			FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
		)`,
		`INSERT INTO donors_new (id, name, blood_type_id, phone, city, created_at, deleted_at)
			SELECT id, name,
				COALESCE((SELECT id FROM blood_types WHERE type = UPPER(TRIM(blood_type))),
					(SELECT id FROM blood_types WHERE type = 'UNKNOWN')),
				phone, city, created_at, deleted_at
			FROM donors`,
		`INSERT INTO recipients_new (id, name, blood_type_id, phone, hospital, created_at, deleted_at)
			SELECT id, name,
				COALESCE((SELECT id FROM blood_types WHERE type = UPPER(TRIM(blood_type))),
					(SELECT id FROM blood_types WHERE type = 'UNKNOWN')),
				phone, hospital, created_at, deleted_at
			FROM recipients`,
		`INSERT INTO donations_new (id, donor_id, units, donation_date, expiry_date, deleted_at)
			SELECT id, donor_id, units, donation_date, expiry_date, deleted_at FROM donations`,
		`INSERT INTO requests_new (id, recipient_id, units, status, request_date, deleted_at)
			SELECT id, recipient_id, units, status, request_date, deleted_at FROM requests`,
		`INSERT INTO inventory_new (id, blood_type_id, units, deleted_at)
			SELECT id,
				COALESCE((SELECT id FROM blood_types WHERE type = UPPER(TRIM(blood_type))),
					(SELECT id FROM blood_types WHERE type = 'UNKNOWN')),
				units, deleted_at
			FROM inventory`,
		`DROP TABLE donors`,
		`DROP TABLE recipients`,
		`DROP TABLE donations`,
		`DROP TABLE requests`,
		`DROP TABLE inventory`,
		`ALTER TABLE donors_new RENAME TO donors`,
		`ALTER TABLE recipients_new RENAME TO recipients`,
		`ALTER TABLE donations_new RENAME TO donations`,
		`ALTER TABLE requests_new RENAME TO requests`,
		`ALTER TABLE inventory_new RENAME TO inventory`,
	}

	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	return err
}
