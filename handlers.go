package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type handler struct {
	db   *sql.DB
	tmpl *template.Template
}

func (h *handler) flash(w http.ResponseWriter, msg string) {
	data, err := loadPageData(h.db, msg)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Println("template error:", err)
	}
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func (h *handler) index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := loadPageData(h.db, "")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Println("template error:", err)
	}
}

func (h *handler) addDonor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	bloodType := normalizeBloodType(r.FormValue("blood_type"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	city := strings.TrimSpace(r.FormValue("city"))
	if name == "" || bloodType == "" {
		h.flash(w, "Donor name and blood type are required.")
		return
	}
	bloodTypeID, err := getOrCreateBloodTypeID(h.db, bloodType)
	if err != nil {
		h.flash(w, "Could not add donor.")
		return
	}
	_, err = h.db.Exec(
		"INSERT INTO donors (name, blood_type_id, phone, city, created_at) VALUES (?, ?, ?, ?, ?)",
		name, bloodTypeID, phone, city, today(),
	)
	if err != nil {
		h.flash(w, "Could not add donor.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) updateDonor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	bloodType := normalizeBloodType(r.FormValue("blood_type"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	city := strings.TrimSpace(r.FormValue("city"))
	if id == 0 || name == "" || bloodType == "" {
		h.flash(w, "Donor update requires id, name, and blood type.")
		return
	}
	bloodTypeID, err := getOrCreateBloodTypeID(h.db, bloodType)
	if err != nil {
		h.flash(w, "Could not update donor.")
		return
	}
	if _, err = h.db.Exec(
		"UPDATE donors SET name = ?, blood_type_id = ?, phone = ?, city = ? WHERE id = ?",
		name, bloodTypeID, phone, city, id,
	); err != nil {
		h.flash(w, "Could not update donor.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) deleteDonor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if _, err := h.db.Exec("UPDATE donors SET deleted_at = ? WHERE id = ?", today(), id); err != nil {
		h.flash(w, "Could not delete donor.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) addRecipient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	bloodType := normalizeBloodType(r.FormValue("blood_type"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	hospital := strings.TrimSpace(r.FormValue("hospital"))
	if name == "" || bloodType == "" {
		h.flash(w, "Recipient name and blood type are required.")
		return
	}
	bloodTypeID, err := getOrCreateBloodTypeID(h.db, bloodType)
	if err != nil {
		h.flash(w, "Could not add recipient.")
		return
	}
	_, err = h.db.Exec(
		"INSERT INTO recipients (name, blood_type_id, phone, hospital, created_at) VALUES (?, ?, ?, ?, ?)",
		name, bloodTypeID, phone, hospital, today(),
	)
	if err != nil {
		h.flash(w, "Could not add recipient.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) updateRecipient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	bloodType := normalizeBloodType(r.FormValue("blood_type"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	hospital := strings.TrimSpace(r.FormValue("hospital"))
	if id == 0 || name == "" || bloodType == "" {
		h.flash(w, "Recipient update requires id, name, and blood type.")
		return
	}
	bloodTypeID, err := getOrCreateBloodTypeID(h.db, bloodType)
	if err != nil {
		h.flash(w, "Could not update recipient.")
		return
	}
	if _, err = h.db.Exec(
		"UPDATE recipients SET name = ?, blood_type_id = ?, phone = ?, hospital = ? WHERE id = ?",
		name, bloodTypeID, phone, hospital, id,
	); err != nil {
		h.flash(w, "Could not update recipient.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) deleteRecipient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if _, err := h.db.Exec("UPDATE recipients SET deleted_at = ? WHERE id = ?", today(), id); err != nil {
		h.flash(w, "Could not delete recipient.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) addDonation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	donorID, _ := strconv.Atoi(r.FormValue("donor_id"))
	units, _ := strconv.Atoi(r.FormValue("units"))
	expiry := strings.TrimSpace(r.FormValue("expiry_date"))
	if donorID == 0 || units <= 0 || expiry == "" {
		h.flash(w, "Donation requires donor, units, and expiry date.")
		return
	}
	bloodTypeID, err := getDonorBloodTypeID(h.db, donorID)
	if err != nil {
		h.flash(w, "Donation requires a valid donor with blood type.")
		return
	}
	tx, err := h.db.Begin()
	if err != nil {
		h.flash(w, "Could not add donation.")
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec(
		"INSERT INTO donations (donor_id, units, donation_date, expiry_date) VALUES (?, ?, ?, ?)",
		donorID, units, today(), expiry,
	); err != nil {
		h.flash(w, "Could not add donation.")
		return
	}
	if err := upsertInventoryByTypeID(tx, bloodTypeID, units); err != nil {
		h.flash(w, "Could not add donation.")
		return
	}
	if err := tx.Commit(); err != nil {
		h.flash(w, "Could not add donation.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) deleteDonation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var units, bloodTypeID int
	err := h.db.QueryRow(`
		SELECT donors.blood_type_id, d.units
		FROM donations d
		JOIN donors ON donors.id = d.donor_id
		WHERE d.id = ? AND d.deleted_at IS NULL
	`, id).Scan(&bloodTypeID, &units)
	if err != nil {
		h.flash(w, "Donation not found.")
		return
	}
	tx, err := h.db.Begin()
	if err != nil {
		h.flash(w, "Could not delete donation.")
		return
	}
	defer tx.Rollback()
	ok, err := consumeInventoryByTypeID(tx, bloodTypeID, units)
	if err != nil {
		h.flash(w, "Inventory update failed.")
		return
	}
	if !ok {
		h.flash(w, "Cannot delete donation because inventory is already used.")
		return
	}
	if _, err := tx.Exec("UPDATE donations SET deleted_at = ? WHERE id = ?", today(), id); err != nil {
		h.flash(w, "Could not delete donation.")
		return
	}
	if err := tx.Commit(); err != nil {
		h.flash(w, "Could not delete donation.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) addRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	recipientID, _ := strconv.Atoi(r.FormValue("recipient_id"))
	units, _ := strconv.Atoi(r.FormValue("units"))
	if recipientID == 0 || units <= 0 {
		h.flash(w, "Request requires recipient and units.")
		return
	}
	if _, err := getRecipientBloodTypeID(h.db, recipientID); err != nil {
		h.flash(w, "Request requires a valid recipient with blood type.")
		return
	}
	if _, err := h.db.Exec(
		"INSERT INTO requests (recipient_id, units, status, request_date) VALUES (?, ?, ?, ?)",
		recipientID, units, "Pending", today(),
	); err != nil {
		h.flash(w, "Could not add request.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) updateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	units, _ := strconv.Atoi(r.FormValue("units"))
	status := strings.TrimSpace(r.FormValue("status"))
	if id == 0 || units <= 0 || status == "" {
		h.flash(w, "Request update requires id, units, and status.")
		return
	}
	if status != "Pending" && status != "Fulfilled" && status != "Cancelled" {
		h.flash(w, "Invalid status value.")
		return
	}
	var oldUnits int
	var oldStatus string
	if err := h.db.QueryRow(
		"SELECT units, status FROM requests WHERE id = ? AND deleted_at IS NULL", id,
	).Scan(&oldUnits, &oldStatus); err != nil {
		h.flash(w, "Request not found.")
		return
	}
	if oldStatus == "Fulfilled" {
		if status != "Fulfilled" || oldUnits != units {
			h.flash(w, "Cannot modify a fulfilled request.")
			return
		}
	}
	if oldStatus != "Fulfilled" && status == "Fulfilled" {
		bloodTypeID, err := getRequestBloodTypeID(h.db, id)
		if err != nil {
			h.flash(w, "Request is missing blood type.")
			return
		}
		tx, err := h.db.Begin()
		if err != nil {
			h.flash(w, "Could not update request.")
			return
		}
		defer tx.Rollback()
		ok, err := consumeInventoryByTypeID(tx, bloodTypeID, units)
		if err != nil {
			h.flash(w, "Inventory update failed.")
			return
		}
		if !ok {
			h.flash(w, "Not enough inventory to fulfill request.")
			return
		}
		if _, err := tx.Exec(
			"UPDATE requests SET units = ?, status = ? WHERE id = ?", units, status, id,
		); err != nil {
			h.flash(w, "Could not update request.")
			return
		}
		if err := tx.Commit(); err != nil {
			h.flash(w, "Could not update request.")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if _, err := h.db.Exec(
		"UPDATE requests SET units = ?, status = ? WHERE id = ?", units, status, id,
	); err != nil {
		h.flash(w, "Could not update request.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) deleteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var status string
	if err := h.db.QueryRow(
		"SELECT status FROM requests WHERE id = ? AND deleted_at IS NULL", id,
	).Scan(&status); err != nil {
		h.flash(w, "Request not found.")
		return
	}
	if status == "Fulfilled" {
		h.flash(w, "Cannot delete a fulfilled request.")
		return
	}
	if _, err := h.db.Exec(
		"UPDATE requests SET status = ?, deleted_at = ? WHERE id = ?", "Cancelled", today(), id,
	); err != nil {
		h.flash(w, "Could not delete request.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) fulfillRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var units, bloodTypeID int
	var status string
	err := h.db.QueryRow(`
		SELECT recipients.blood_type_id, r.units, r.status
		FROM requests r
		JOIN recipients ON recipients.id = r.recipient_id
		WHERE r.id = ? AND r.deleted_at IS NULL
	`, id).Scan(&bloodTypeID, &units, &status)
	if err != nil {
		h.flash(w, "Request not found.")
		return
	}
	if status != "Pending" {
		h.flash(w, "Only pending requests can be fulfilled.")
		return
	}
	tx, err := h.db.Begin()
	if err != nil {
		h.flash(w, "Could not fulfill request.")
		return
	}
	defer tx.Rollback()
	ok, err := consumeInventoryByTypeID(tx, bloodTypeID, units)
	if err != nil {
		h.flash(w, "Inventory update failed.")
		return
	}
	if !ok {
		h.flash(w, "Not enough inventory to fulfill request.")
		return
	}
	if _, err := tx.Exec("UPDATE requests SET status = ? WHERE id = ?", "Fulfilled", id); err != nil {
		h.flash(w, "Could not update request.")
		return
	}
	if err := tx.Commit(); err != nil {
		h.flash(w, "Could not fulfill request.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
