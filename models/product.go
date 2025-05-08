package models

import (
	"database/sql"
	"ecommerce-app/db"
	"fmt"
)

// Product represents an item available for purchase
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

// FormatPrice returns a formatted price string
func (p *Product) FormatPrice() string {
	return fmt.Sprintf("$%.2f", p.Price)
}

// SeedProducts inserts initial product data into the database.
func SeedProducts() error {
	products := []Product{
		{
			ID:          "prod_1",
			Name:        "Premium T-Shirt",
			Description: "High quality cotton t-shirt with logo",
			Price:       29.99,
			ImageURL:    "/static/img/placeholder.svg",
		},
		{
			ID:          "prod_2",
			Name:        "Designer Jeans",
			Description: "Comfortable jeans for everyday wear",
			Price:       89.99,
			ImageURL:    "/static/img/placeholder.svg",
		},
		{
			ID:          "prod_3",
			Name:        "Running Shoes",
			Description: "Lightweight shoes for optimal performance",
			Price:       119.99,
			ImageURL:    "/static/img/placeholder.svg",
		},
	}

	for _, p := range products {
		// Check if product already exists
		var count int
		row := db.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", p.ID)
		err := row.Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error checking product existence: %w", err)
		}

		if count == 0 {
			// Insert product if it doesn't exist
			_, err := db.DB.Exec(
				"INSERT INTO products (id, name, description, price, image_url) VALUES (?, ?, ?, ?, ?)",
				p.ID, p.Name, p.Description, p.Price, p.ImageURL,
			)
			if err != nil {
				return fmt.Errorf("error inserting product %s: %w", p.ID, err)
			}
		}
	}
	return nil
}

// GetProducts returns a list of all products from the database.
func GetProducts() ([]Product, error) {
	rows, err := db.DB.Query("SELECT id, name, description, price, image_url FROM products")
	if err != nil {
		return nil, fmt.Errorf("error fetching products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL); err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating through product rows: %w", err)
	}

	return products, nil
}

// GetProductByID returns a product with the specified ID from the database.
func GetProductByID(id string) (Product, error) {
	row := db.DB.QueryRow("SELECT id, name, description, price, image_url FROM products WHERE id = ?", id)

	var p Product
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL)
	if err == sql.ErrNoRows {
		return Product{}, fmt.Errorf("product with ID %s not found", id)
	}
	if err != nil {
		return Product{}, fmt.Errorf("error fetching product by ID %s: %w", id, err)
	}

	return p, nil
}
