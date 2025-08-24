package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB veritabanını başlatır ve gerekli tabloları oluşturur
func InitDB() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./agri_management.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Veritabanı bağlantısını test et
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Tabloları oluştur
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("✅ Veritabanı başarıyla başlatıldı")
	return db, nil
}

// createTables gerekli tabloları oluşturur
func createTables(db *sql.DB) error {
	tables := []string{
		createUsersTable,
		createLandsTable,
		createLivestockTable,
		createProductionTable,
		createTransactionsTable,
		createEventsTable,
		createNotificationsTable,
		createHealthRecordsTable,
		createMilkProductionTable,
		createLandActivitiesTable,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return err
		}
	}

	log.Println("✅ Tüm tablolar başarıyla oluşturuldu")
	return nil
}

// Tablo oluşturma SQL komutları
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    avatar TEXT,
    role TEXT DEFAULT 'farmer',
    farm_name TEXT,
    location TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createLandsTable = `
CREATE TABLE IF NOT EXISTS lands (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    area REAL NOT NULL,
    unit TEXT NOT NULL,
    crop TEXT,
    status TEXT DEFAULT 'active',
    last_activity DATETIME,
    productivity REAL DEFAULT 0,
    latitude REAL,
    longitude REAL,
    address TEXT,
    soil_type TEXT,
    irrigation_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createLivestockTable = `
CREATE TABLE IF NOT EXISTS livestock (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tag_number TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    breed TEXT,
    gender TEXT,
    birth_date DATE,
    weight REAL,
    health_status TEXT DEFAULT 'healthy',
    location TEXT,
    mother TEXT,
    father TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createProductionTable = `
CREATE TABLE IF NOT EXISTS production (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    land_id TEXT,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    amount REAL NOT NULL,
    unit TEXT NOT NULL,
    harvest_date DATE,
    quality TEXT,
    storage_location TEXT,
    status TEXT DEFAULT 'active',
    price REAL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (land_id) REFERENCES lands(id) ON DELETE SET NULL
);`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    amount REAL NOT NULL,
    currency TEXT DEFAULT 'TRY',
    date DATE NOT NULL,
    status TEXT DEFAULT 'completed',
    payment_method TEXT,
    receipt TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createEventsTable = `
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    is_all_day BOOLEAN DEFAULT FALSE,
    status TEXT DEFAULT 'pending',
    priority TEXT DEFAULT 'medium',
    location TEXT,
    related_entity_type TEXT,
    related_entity_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createNotificationsTable = `
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL,
    priority TEXT DEFAULT 'medium',
    is_read BOOLEAN DEFAULT FALSE,
    related_entity_type TEXT,
    related_entity_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createHealthRecordsTable = `
CREATE TABLE IF NOT EXISTS health_records (
    id TEXT PRIMARY KEY,
    livestock_id TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    date DATE NOT NULL,
    veterinarian TEXT,
    cost REAL,
    notes TEXT,
    next_checkup DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (livestock_id) REFERENCES livestock(id) ON DELETE CASCADE
);`

const createMilkProductionTable = `
CREATE TABLE IF NOT EXISTS milk_production (
    id TEXT PRIMARY KEY,
    livestock_id TEXT NOT NULL,
    date DATE NOT NULL,
    amount REAL NOT NULL,
    quality TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (livestock_id) REFERENCES livestock(id) ON DELETE CASCADE
);`

const createLandActivitiesTable = `
CREATE TABLE IF NOT EXISTS land_activities (
    id TEXT PRIMARY KEY,
    land_id TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    scheduled_date DATE,
    actual_date DATE,
    notes TEXT,
    cost REAL,
    result TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (land_id) REFERENCES lands(id) ON DELETE CASCADE
);`
