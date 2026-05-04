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
-- ==============================================================================
-- Blood Bank Management System - PL/SQL Implementations
-- ==============================================================================
-- Note: This file demonstrates how the core business logic of the system 
-- would be implemented using Oracle PL/SQL (Stored Procedures, Triggers, 
-- Functions, and Cursors). The actual application uses Go and SQLite, 
-- but these scripts fulfill DBMS project requirements for PL/SQL implementation.
-- ==============================================================================

SET SERVEROUTPUT ON;

-- ------------------------------------------------------------------------------
-- 1. TRIGGER: Auto-Update Inventory on New Donation
-- ------------------------------------------------------------------------------
-- Automatically increases the inventory units for the corresponding blood type
-- whenever a new donation is successfully recorded in the 'donations' table.
-- ------------------------------------------------------------------------------
CREATE OR REPLACE TRIGGER trg_after_donation_insert
AFTER INSERT ON donations
FOR EACH ROW
DECLARE
    v_blood_type_id donors.blood_type_id%TYPE;
    v_inventory_count NUMBER;
BEGIN
    -- Get the blood type of the donor
    SELECT blood_type_id INTO v_blood_type_id
    FROM donors
    WHERE id = :NEW.donor_id;

    -- Check if an inventory record already exists for this blood type
    SELECT COUNT(*) INTO v_inventory_count
    FROM inventory
    WHERE blood_type_id = v_blood_type_id;

    IF v_inventory_count > 0 THEN
        -- Update existing inventory
        UPDATE inventory
        SET units = units + :NEW.units
        WHERE blood_type_id = v_blood_type_id;
    ELSE
        -- Insert new inventory record
        INSERT INTO inventory (blood_type_id, units)
        VALUES (v_blood_type_id, :NEW.units);
    END IF;
    
    DBMS_OUTPUT.PUT_LINE('Inventory updated successfully for donation.');
END;
/

-- ------------------------------------------------------------------------------
-- 2. PROCEDURE: Fulfill Blood Request
-- ------------------------------------------------------------------------------
-- Takes a request ID, checks if there is sufficient stock in the inventory.
-- If stock is sufficient, deducts the requested units and marks the request 
-- as 'Fulfilled'. If insufficient, it raises an application error to rollback.
-- ------------------------------------------------------------------------------
CREATE OR REPLACE PROCEDURE fulfill_blood_request (
    p_request_id IN requests.id%TYPE
)
IS
    v_recipient_id requests.recipient_id%TYPE;
    v_requested_units requests.units%TYPE;
    v_status requests.status%TYPE;
    v_blood_type_id recipients.blood_type_id%TYPE;
    v_available_units inventory.units%TYPE;
BEGIN
    -- Fetch request details
    SELECT recipient_id, units, status 
    INTO v_recipient_id, v_requested_units, v_status
    FROM requests 
    WHERE id = p_request_id AND deleted_at IS NULL;

    -- Ensure request is not already fulfilled or cancelled
    IF v_status != 'Pending' THEN
        RAISE_APPLICATION_ERROR(-20001, 'Request is already fulfilled or cancelled.');
    END IF;

    -- Get recipient's blood type
    SELECT blood_type_id INTO v_blood_type_id
    FROM recipients
    WHERE id = v_recipient_id;

    -- Check available inventory (handle case where no inventory row exists yet)
    BEGIN
        SELECT units INTO v_available_units
        FROM inventory
        WHERE blood_type_id = v_blood_type_id AND deleted_at IS NULL;
    EXCEPTION
        WHEN NO_DATA_FOUND THEN
            v_available_units := 0;
    END;

    -- Fulfill if sufficient stock exists
    IF v_available_units >= v_requested_units THEN
        -- Deduct from inventory
        UPDATE inventory
        SET units = units - v_requested_units
        WHERE blood_type_id = v_blood_type_id;

        -- Update request status
        UPDATE requests
        SET status = 'Fulfilled'
        WHERE id = p_request_id;

        COMMIT;
        DBMS_OUTPUT.PUT_LINE('Request fulfilled successfully.');
    ELSE
        RAISE_APPLICATION_ERROR(-20002, 'Insufficient blood inventory to fulfill this request.');
    END IF;

EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RAISE_APPLICATION_ERROR(-20003, 'Request ID not found or is deleted.');
    WHEN OTHERS THEN
        ROLLBACK;
        RAISE;
END fulfill_blood_request;
/

-- ------------------------------------------------------------------------------
-- 3. PROCEDURE WITH CURSOR: Generate Pending Requests Report
-- ------------------------------------------------------------------------------
-- Uses a cursor to iterate through all 'Pending' blood requests and prints
-- a formatted report joining across multiple tables.
-- ------------------------------------------------------------------------------
CREATE OR REPLACE PROCEDURE print_pending_requests
IS
    -- Define the cursor to fetch pending requests
    CURSOR c_pending IS
        SELECT r.id, rec.name AS recipient_name, bt.type AS blood_type, r.units, r.request_date
        FROM requests r
        JOIN recipients rec ON r.recipient_id = rec.id
        JOIN blood_types bt ON rec.blood_type_id = bt.id
        WHERE r.status = 'Pending' AND r.deleted_at IS NULL;
        
    v_count NUMBER := 0;
BEGIN
    DBMS_OUTPUT.PUT_LINE('--------------------------------------------------');
    DBMS_OUTPUT.PUT_LINE('           PENDING BLOOD REQUESTS REPORT          ');
    DBMS_OUTPUT.PUT_LINE('--------------------------------------------------');
    
    -- Iterate through the cursor
    FOR req IN c_pending LOOP
        DBMS_OUTPUT.PUT_LINE('Req ID: ' || req.id || 
                             ' | Recipient: ' || RPAD(req.recipient_name, 20) || 
                             ' | Type: ' || RPAD(req.blood_type, 3) || 
                             ' | Units: ' || req.units || 
                             ' | Date: ' || req.request_date);
        v_count := v_count + 1;
    END LOOP;
    
    DBMS_OUTPUT.PUT_LINE('--------------------------------------------------');
    IF v_count = 0 THEN
        DBMS_OUTPUT.PUT_LINE('No pending requests found.');
    ELSE
        DBMS_OUTPUT.PUT_LINE('Total Pending Requests: ' || v_count);
    END IF;
END print_pending_requests;
/

-- ------------------------------------------------------------------------------
-- 4. FUNCTION: Get Current Inventory Level
-- ------------------------------------------------------------------------------
-- Returns the total units available for a specific blood type string (e.g., 'O+').
-- ------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION get_inventory_level (
    p_blood_type IN blood_types.type%TYPE
) RETURN NUMBER
IS
    v_units NUMBER;
BEGIN
    SELECT COALESCE(i.units, 0) INTO v_units
    FROM blood_types bt
    LEFT JOIN inventory i ON bt.id = i.blood_type_id
    WHERE bt.type = p_blood_type
      AND i.deleted_at IS NULL;
      
    RETURN v_units;
EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RETURN 0;
END get_inventory_level;
/
