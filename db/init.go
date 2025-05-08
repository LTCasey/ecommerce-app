// DB init stub

package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // Import the pure Go SQLite driver
)

var DB *sql.DB

// InitDB initializes the database connection and creates tables.
func InitDB() {
	var err error
	// Use an in-memory database for simplicity. For persistent storage, use a file path like "ecommerce.db"
	DB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create products table
	createProductsTableSQL := `CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		image_url TEXT
	);`

	_, err = DB.Exec(createProductsTableSQL)
	if err != nil {
		log.Fatalf("Error creating products table: %v", err)
	}

	// Create orders table
	createOrdersTableSQL := `CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		customer_email TEXT NOT NULL,
		total_amount REAL NOT NULL,
		status TEXT NOT NULL,
		stripe_id TEXT,
		created_at DATETIME,
		updated_at DATETIME
	);`

	_, err = DB.Exec(createOrdersTableSQL)
	if err != nil {
		log.Fatalf("Error creating orders table: %v", err)
	}

	// Create order_items table
	createOrderItemsTableSQL := `CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id TEXT NOT NULL,
		product_id TEXT NOT NULL,
		product_name TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		unit_price REAL NOT NULL,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
	);`

	_, err = DB.Exec(createOrderItemsTableSQL)
	if err != nil {
		log.Fatalf("Error creating order_items table: %v", err)
	}

	log.Println("Database initialized and tables created.")
}
