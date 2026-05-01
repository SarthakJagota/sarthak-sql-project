package main

type Donor struct {
	ID        int
	Name      string
	BloodType string
	Phone     string
	City      string
	CreatedAt string
}

type Recipient struct {
	ID        int
	Name      string
	BloodType string
	Phone     string
	Hospital  string
	CreatedAt string
}

type Donation struct {
	ID           int
	DonorID      int
	DonorName    string
	BloodType    string
	Units        int
	DonationDate string
	ExpiryDate   string
}

type Inventory struct {
	BloodType  string
	DonorName  string
	Units      int
	ExpiryDate string
}

type Request struct {
	ID          int
	RecipientID int
	Recipient   string
	BloodType   string
	Units       int
	Status      string
	RequestDate string
}

type PageData struct {
	Donors     []Donor
	Recipients []Recipient
	Donations  []Donation
	Inventory  []Inventory
	Requests   []Request
	Message    string
}
