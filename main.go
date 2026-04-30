package main

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

//go:embed templates/* static/* sql/*
var assets embed.FS

func main() {
	db, err := sql.Open("sqlite", "file:bloodbank.db?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFS(assets, "templates/index.html"))
	h := &handler{db: db, tmpl: tmpl}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(assets)))
	mux.HandleFunc("/", h.index)
	mux.HandleFunc("/donors", h.addDonor)
	mux.HandleFunc("/donors/update", h.updateDonor)
	mux.HandleFunc("/donors/delete", h.deleteDonor)
	mux.HandleFunc("/recipients", h.addRecipient)
	mux.HandleFunc("/recipients/update", h.updateRecipient)
	mux.HandleFunc("/recipients/delete", h.deleteRecipient)
	mux.HandleFunc("/donations", h.addDonation)
	mux.HandleFunc("/donations/delete", h.deleteDonation)
	mux.HandleFunc("/requests", h.addRequest)
	mux.HandleFunc("/requests/update", h.updateRequest)
	mux.HandleFunc("/requests/delete", h.deleteRequest)
	mux.HandleFunc("/fulfill", h.fulfillRequest)

	addr := ":8080"
	log.Println("Blood Bank DBMS running on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
