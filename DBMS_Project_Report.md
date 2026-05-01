# DBMS Project Report: Blood Bank Management System

## 1. Title Page
- **Project Title:** Blood Bank Management System
- **Course Name & Code:** UCS310 – Database Management Systems
- **Degree & Year:** B.Tech (2nd Year)
- **Department / Institute Name:** [To Be Filled] / [To Be Filled]
- **Group Members:** 
  1. [Name 1] (Roll No: [Roll 1])
  2. [Name 2] (Roll No: [Roll 2])
  3. [Name 3] (Roll No: [Roll 3])
- **Lab Instructor Name:** [To Be Filled]
- **Academic Year:** 2025–26

---

## 2. Introduction
The Blood Bank Management System is a database-driven web application designed to manage blood donors, recipients, blood requests, and inventory efficiently. The system stores and manages donor and recipient records, donation logs, request status, and blood stock information in real-time.

Traditionally, many blood banks manage donor records and requests manually or using basic spreadsheets. This file-based approach suffers from data redundancy, inconsistency, delayed updates, and lack of security. A Database Management System (DBMS) resolves these issues by providing structured data storage, integrity constraints (like preventing invalid blood types), and efficient query processing. 

This project emphasizes backend implementation using SQL and relational database concepts, utilizing SQLite as the primary database, alongside advanced PL/SQL implementations for academic demonstration.

---

## 3. Problem Statement
In many hospitals and blood banks, operations such as tracking donations and fulfilling blood requests are handled manually. This leads to:
- Duplicate or outdated donor/recipient records
- Inaccurate inventory and expiry tracking
- Difficulty in matching blood requests to available stock in emergency situations
- Slow data retrieval and reporting
- Lack of data integrity and audit trails

The proposed Blood Bank Management System provides a structured relational database solution to manage donors, donations, requests, and inventory efficiently, ensuring reliable and secure operations.

---

## 4. Objectives of the Project
- To design a database using an Entity-Relationship (E-R) data model.
- To convert the ER model into relational tables with strict foreign key relationships.
- To apply normalization (up to 3NF) to eliminate data redundancy.
- To implement the database using standard SQL (DDL and DML commands).
- To demonstrate advanced PL/SQL constructs like stored procedures, functions, triggers, and cursors.
- To ensure data consistency using constraints, soft deletes, and transactions (ACID properties).

---

## 5. Scope of the Project
**Functional Boundaries:** The system manages the complete lifecycle of a blood unit, from donor registration and donation recording to recipient registration, request tracking, and inventory fulfillment.
**Types of Users involved:** 
- Admin/Staff (Hospital or Blood Bank personnel)
**Modules Covered:**
- Donor Registration Module
- Recipient Registration Module
- Donation Recording Module
- Blood Request Module
- Real-time Inventory Tracking Module

---

## 6. Proposed System Description
The system is developed using Go (Golang) for backend logic, SQLite for the database, and HTML/CSS for the frontend interface.

**Working of the System:**
1. Admin/staff registers donors and recipients (recording blood types, contact info, etc.).
2. Donations are recorded, capturing the donated units and expiry date, and are linked to the respective donor.
3. Inventory is updated automatically based on the newly recorded donations.
4. Blood requests are created by recipients and tracked by status (Pending, Fulfilled, Cancelled).
5. When a request is fulfilled, the required units are dynamically deducted from the inventory.
6. The system uses "soft deletes" (marking a `deleted_at` timestamp) instead of hard deletes to maintain a complete historical audit trail.

This system greatly improves operational efficiency, data consistency, and traceability of all blood units.

---

## 7. Database Design

### 7.1 Entity–Relationship (ER) Diagram
**Identified Entities:**
1. `Donor`
2. `Donation`
3. `Recipient`
4. `Request`
5. `Inventory`
6. `BloodType` (Lookup table)

**Relationships:**
- A `Donor` submits one or more `Donation`s (1-M).
- A `Recipient` makes one or more `Request`s (1-M).
- `Inventory` summarizes stock for exactly one `BloodType` (1-1).
- `BloodType` is referenced by `Donor`, `Recipient`, and `Inventory` to maintain consistency.

*(Note: ER diagram image `sarthak 3nf.png` is available in the project repository and should be attached here in the final print).*

### 7.2 Relational Schema
**BLOOD_TYPES**
`id` (PK), `type` (UNIQUE)

**DONORS**
`id` (PK), `name`, `blood_type_id` (FK), `phone`, `city`, `created_at`, `deleted_at`

**RECIPIENTS**
`id` (PK), `name`, `blood_type_id` (FK), `phone`, `hospital`, `created_at`, `deleted_at`

**DONATIONS**
`id` (PK), `donor_id` (FK), `units`, `donation_date`, `expiry_date`, `deleted_at`

**REQUESTS**
`id` (PK), `recipient_id` (FK), `units`, `status`, `request_date`, `deleted_at`

**INVENTORY**
`id` (PK), `blood_type_id` (FK, UNIQUE), `units`, `deleted_at`

---

## 8. Normalization

**Functional Dependencies:**
- `Donors:` id → name, blood_type_id, phone, city, created_at
- `BloodTypes:` id → type
- `Donations:` id → donor_id, units, donation_date, expiry_date
- `Recipients:` id → name, blood_type_id, phone, hospital, created_at
- `Requests:` id → recipient_id, units, status, request_date
- `Inventory:` blood_type_id → units

**1NF (First Normal Form)**
- All attributes are atomic (single-valued).
- Each table has a primary key and no repeating groups.

**2NF (Second Normal Form)**
- All tables use single-column primary keys, ensuring no partial dependency exists.
- Every non-key attribute depends fully on its table's primary key.

**3NF (Third Normal Form)**
- Blood type values are stored once in the `BLOOD_TYPES` lookup table and referenced via foreign keys (`blood_type_id`).
- Redundant blood type text fields in donations, requests, and inventory tables are removed.
- All non-key attributes depend only on the primary key, eliminating transitive dependencies.

**Conclusion:** The database is successfully normalized up to 3NF.

---

## 9. Database Implementation

### 9.1 SQL Implementation
**DDL Commands (Table Creation):**
```sql
CREATE TABLE IF NOT EXISTS blood_types (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS donors (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT NOT NULL,
	blood_type_id INTEGER NOT NULL,
	FOREIGN KEY(blood_type_id) REFERENCES blood_types(id)
);

CREATE TABLE IF NOT EXISTS donations (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	donor_id      INTEGER NOT NULL,
	units         INTEGER NOT NULL,
	donation_date TEXT NOT NULL,
	FOREIGN KEY(donor_id) REFERENCES donors(id)
);
```

**DML Commands (Insert, Update, Select):**
```sql
-- INSERT Example
INSERT INTO requests (recipient_id, units, status, request_date) 
VALUES (1, 2, 'Pending', '2026-05-01');

-- UPDATE Example (Soft Delete)
UPDATE donors SET deleted_at = '2026-05-01' WHERE id = 5;

-- SELECT Query with JOIN Example
SELECT d.id, d.name, bt.type, d.phone 
FROM donors d 
JOIN blood_types bt ON d.blood_type_id = bt.id 
WHERE d.deleted_at IS NULL;
```

### 9.2 PL/SQL Components
While the production application utilizes Go for backend logic, the following PL/SQL components have been implemented (available in `sql/plsql_implementation.sql`) to demonstrate advanced DBMS functionalities:

- **Triggers:** `trg_after_donation_insert` automatically updates the `inventory` table whenever a new row is added to the `donations` table.
- **Stored Procedures:** `fulfill_blood_request` encapsulates the logic of checking inventory stock, deducting units, and updating the request status to 'Fulfilled', raising exceptions if stock is insufficient.
- **Functions:** `get_inventory_level` returns the available stock as a number for a given blood type string (e.g., 'O+').
- **Cursors:** `print_pending_requests` uses a cursor to iterate over a multi-table JOIN result of all 'Pending' requests to generate a formatted report.

---

## 10. Transaction Management & Concurrency
The system heavily utilizes Database Transactions to maintain ACID properties (Atomicity, Consistency, Isolation, Durability), particularly during complex operations like fulfilling a blood request or recording a donation.

**Implementation Example:**
When fulfilling a request, the system must (1) deduct units from inventory, and (2) update the request status. 
```go
tx, err := db.Begin()
// ... Deduct inventory ...
// ... Update request status ...
tx.Commit() // Persist changes if both succeed
// tx.Rollback() is called if any error occurs to prevent partial updates.
```
This ensures that if the system crashes midway, blood inventory is never deducted without the request also being marked as fulfilled.

---

## 11. Tools & Technologies Used
- **DBMS Software:** SQLite 3 (with standard SQL logic easily portable to Oracle/PostgreSQL)
- **Programming Languages:** Go (Golang) 1.22, SQL, PL/SQL
- **Frontend:** HTML5, CSS3 (Vanilla), JavaScript

---

## 12. Expected Outcomes
- Successful creation of a fully functional, normalized database for a blood bank.
- Efficient data retrieval and management using SQL joins and queries.
- Automation of inventory updates using database triggers and transactions.
- Zero data loss through the implementation of soft deletes.
- High data consistency and integrity through the enforcement of foreign key constraints and 3NF normalization.
