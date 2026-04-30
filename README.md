# Blood Bank Management System

A full-stack Blood Bank Management System built with **Go**, **SQLite**, and a plain HTML/CSS frontend. Developed as a DBMS course project (UCS310) demonstrating relational database design, normalization (1NF → 3NF), and a complete web application workflow.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Database Design](#database-design)
  - [Schema (3NF)](#schema-3nf)
  - [Normalization](#normalization)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation & Run](#installation--run)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Implementation Highlights](#implementation-highlights)

---

## Overview

This system allows hospital/blood bank staff to:

- Register and manage **blood donors** and **recipients**
- Record **blood donations** and track expiry dates
- Maintain a real-time **inventory** of blood units per blood type
- Create and fulfill **blood requests**, with automatic inventory deduction
- Perform soft-delete operations to preserve audit history

---

## Features

| Feature | Description |
|---|---|
| Donor Management | Add, update, and soft-delete donor records |
| Recipient Management | Add, update, and soft-delete recipient records |
| Donation Tracking | Log donations with units and expiry date; inventory auto-updates |
| Blood Requests | Create requests, track status (Pending / Fulfilled / Cancelled) |
| Inventory Management | Real-time stock per blood type; auto-deducted on fulfillment |
| 3NF Database | Normalized schema with a `blood_types` lookup table and foreign keys |
| Auto-Migration | On startup, migrates legacy denormalized data to the 3NF schema |
| Embedded Assets | Templates and CSS compiled into the binary via Go `embed` |
| Soft Deletes | Records are never physically deleted; `deleted_at` timestamp is set instead |

---

## Tech Stack

**Backend**
- [Go 1.22](https://go.dev/) — HTTP server, template rendering, business logic
- [modernc.org/sqlite v1.30.0](https://pkg.go.dev/modernc.org/sqlite) — Pure-Go SQLite driver (no CGo required)
- Go standard library: `net/http`, `database/sql`, `html/template`, `embed`

**Frontend**
- HTML5 + CSS3 (dark theme, CSS variables, Grid/Flexbox)
- Vanilla JavaScript (blood-type auto-sync between dropdowns)
- [Google Fonts](https://fonts.google.com/) — Bebas Neue, Space Grotesk

**Database**
- SQLite 3 (file: `bloodbank.db`, auto-created on first run)

---

## Database Design

### Schema (3NF)

```
blood_types          donors                  recipients
───────────          ──────                  ──────────
id (PK)              id (PK)                 id (PK)
type (UNIQUE)        name                    name
                     blood_type_id (FK)      blood_type_id (FK)
                     phone                   phone
                     city                    hospital
                     created_at              created_at
                     deleted_at              deleted_at

donations            requests                inventory
─────────            ────────                ─────────
id (PK)              id (PK)                 id (PK)
donor_id (FK)        recipient_id (FK)       blood_type_id (FK, UNIQUE)
units                units                   units
donation_date        status                  deleted_at
expiry_date          request_date
deleted_at           deleted_at
```

All foreign keys are enforced via `PRAGMA foreign_keys = ON`.

### Normalization

| Form | Key Rule Applied |
|---|---|
| **1NF** | All columns are atomic; no repeating groups |
| **2NF** | No partial dependencies (all non-key columns depend on the full PK) |
| **3NF** | `blood_type` text moved to a separate `blood_types` lookup table; all tables reference it by `blood_type_id` FK, eliminating transitive dependencies |

Schema files for each normal form are available in the repo root:
- [schema_raw.dbml](schema_raw.dbml) — Denormalized (raw)
- [schema_1nf.dbml](schema_1nf.dbml) — 1NF
- [schema_2nf.dbml](schema_2nf.dbml) — 2NF
- [schema_3nf.dbml](schema_3nf.dbml) — 3NF (final)

ER diagrams: [sarthak raw.png](sarthak%20raw.png) and [sarthak 3nf.png](sarthak%203nf.png)

---

## Project Structure

```
sarthak-sql-project/
├── main.go                  # Entry point — server setup and route registration
├── handlers.go              # HTTP handler methods (handler struct)
├── db.go                    # Database layer — init, migration, queries, load functions
├── models.go                # Struct types (Donor, Recipient, Donation, etc.)
├── go.mod                   # Go module definition
├── go.sum                   # Dependency lock file
├── bloodbank.db             # SQLite database (auto-created at runtime, gitignored)
├── sql/
│   └── schema.sql           # DDL — CREATE TABLE statements for all 6 tables
├── templates/
│   └── index.html           # Single-page HTML template (forms + tables)
├── static/
│   └── style.css            # Dark-themed responsive stylesheet
├── schema_raw.dbml          # Denormalized schema (DBML format)
├── schema_1nf.dbml          # 1NF schema
├── schema_2nf.dbml          # 2NF schema
├── schema_3nf.dbml          # 3NF schema (matches production)
├── sarthak raw.png          # ER diagram — raw schema
├── sarthak 3nf.png          # ER diagram — 3NF schema
├── SYNOPSIS.md              # Full academic project documentation
└── README.md                # This file
```

---

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/) installed and on your `PATH`
- No other dependencies — SQLite is bundled as a pure-Go package

### Installation & Run

```bash
# 1. Clone the repository
git clone <repo-url>
cd sarthak-sql-project

# 2. Download Go dependencies
go mod tidy

# 3. Start the server
go run .
```

The server starts on **http://localhost:8080**.  
The database file `bloodbank.db` is created automatically in the project root on first run.

---

## Usage

Open **http://localhost:8080** in your browser. The single-page interface is divided into sections:

### 1. Donors
- **Add Donor** — Fill in name, blood type, phone (optional), city (optional)
- **Edit** — Click the edit button on any row to update details inline
- **Delete** — Soft-deletes the record (sets `deleted_at`; data is not lost)

### 2. Recipients
- **Add Recipient** — Fill in name, blood type, phone (optional), hospital (optional)
- Same edit/delete workflow as donors

### 3. Record a Donation
- Select a donor from the dropdown (blood type auto-fills)
- Enter units donated and expiry date
- On submit, a `donations` record is created and `inventory` is automatically incremented for that blood type

### 4. Blood Requests
- Select a recipient (blood type auto-fills)
- Enter units requested
- New requests start as **Pending**
- Click **Fulfill** to deduct units from inventory and mark the request as **Fulfilled**
- Click **Cancel** to mark as **Cancelled** without touching inventory

### 5. Inventory
- Read-only view of current stock per blood type
- Updates automatically after every donation or request fulfillment

---

## API Endpoints

Routes are registered in [main.go](main.go); handlers live in [handlers.go](handlers.go).

| Method | Path | Handler | Description |
|---|---|---|---|
| `GET` | `/` | `index` | Render the main dashboard |
| `POST` | `/donors` | `addDonor` | Add a new donor |
| `POST` | `/donors/update` | `updateDonor` | Update donor details |
| `POST` | `/donors/delete` | `deleteDonor` | Soft-delete a donor |
| `POST` | `/recipients` | `addRecipient` | Add a new recipient |
| `POST` | `/recipients/update` | `updateRecipient` | Update recipient details |
| `POST` | `/recipients/delete` | `deleteRecipient` | Soft-delete a recipient |
| `POST` | `/donations` | `addDonation` | Record a donation (also updates inventory) |
| `POST` | `/donations/delete` | `deleteDonation` | Soft-delete a donation record |
| `POST` | `/requests` | `addRequest` | Create a blood request |
| `POST` | `/requests/update` | `updateRequest` | Update request status |
| `POST` | `/requests/delete` | `deleteRequest` | Soft-delete a request |
| `POST` | `/fulfill` | `fulfillRequest` | Fulfill a request (deducts inventory) |

---

## Implementation Highlights

**Inventory Upsert**  
When a donation is recorded, the inventory row for that blood type is updated if it already exists, or inserted if not — using SQLite's `INSERT OR REPLACE` / upsert logic. This keeps inventory consistent without manual initialization.

**Soft Deletes**  
No `DELETE` SQL statements are executed. Every "delete" sets `deleted_at = CURRENT_TIMESTAMP`. All `SELECT` queries include `WHERE deleted_at IS NULL`, so deleted records are invisible to the UI but retained in the database for auditing.

**Auto-Migration**  
On startup, `migrateTo3NF()` checks whether the schema is still on the legacy denormalized format and, if so, migrates data to the normalized 3NF structure automatically — preserving all existing records.

**Embedded Assets**  
The `//go:embed` directive bundles `templates/`, `static/`, and `sql/` into the compiled binary. `db.go` reads `sql/schema.sql` at startup via `assets.ReadFile`. The result is a single self-contained executable with no external file dependencies.

**Foreign Key Enforcement**  
SQLite foreign keys are disabled by default. The application explicitly runs `PRAGMA foreign_keys = ON` for every new connection to enforce referential integrity.
