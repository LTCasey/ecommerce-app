package models

import (
	"database/sql"
	"ecommerce-app/db"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // Import the uuid package
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	// OrderStatusPending means payment is in progress
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusCompleted means payment was successful
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusFailed means payment failed
	OrderStatusFailed OrderStatus = "failed"
)

// OrderItem represents a product in an order
type OrderItem struct {
	ID          int     `json:"id"` // Database ID for order item
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Order represents a customer purchase
type Order struct {
	ID            string      `json:"id"` // UUID for the order
	CustomerEmail string      `json:"customer_email"`
	Items         []OrderItem `json:"items"`
	TotalAmount   float64     `json:"total_amount"`
	Status        OrderStatus `json:"status"`
	StripeID      string      `json:"stripe_id"` // Stripe Checkout Session ID
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// CalculateTotal calculates the total amount for the order
func (o *Order) CalculateTotal() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}
	o.TotalAmount = total
	return total
}

// NewOrder creates a new order with initial values
func NewOrder(customerEmail string) *Order {
	return &Order{
		ID:            generateOrderID(), // Use UUID
		CustomerEmail: customerEmail,
		Items:         []OrderItem{},
		Status:        OrderStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// AddItem adds a product to the order
func (o *Order) AddItem(product Product, quantity int) {
	// Check if product already exists in order
	for i, item := range o.Items {
		if item.ProductID == product.ID {
			o.Items[i].Quantity += quantity
			o.CalculateTotal()
			return
		}
	}

	// Add new item
	o.Items = append(o.Items, OrderItem{
		ProductID:   product.ID,
		ProductName: product.Name,
		Quantity:    quantity,
		UnitPrice:   product.Price,
	})
	o.CalculateTotal()
}

// generateOrderID creates a UUID for the order ID
func generateOrderID() string {
	return uuid.New().String()
}

// Save saves the order and its items to the database.
func (o *Order) Save() error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if commit fails

	// Insert or update order
	_, err = tx.Exec(
		"INSERT INTO orders (id, customer_email, total_amount, status, stripe_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET customer_email=excluded.customer_email, total_amount=excluded.total_amount, status=excluded.status, stripe_id=excluded.stripe_id, updated_at=excluded.updated_at",
		o.ID, o.CustomerEmail, o.TotalAmount, string(o.Status), o.StripeID, o.CreatedAt, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error saving order %s: %w", o.ID, err)
	}

	// Delete existing order items for this order (simpler for now, could optimize)
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", o.ID)
	if err != nil {
		return fmt.Errorf("error deleting existing order items for order %s: %w", o.ID, err)
	}

	// Insert order items
	for _, item := range o.Items {
		_, err := tx.Exec(
			"INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price) VALUES (?, ?, ?, ?, ?)",
			o.ID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("error saving order item for order %s: %w", o.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("Order %s saved to database.", o.ID)
	return nil
}

// GetOrderByStripeID retrieves an order by its Stripe Checkout Session ID.
func GetOrderByStripeID(stripeID string) (*Order, error) {
	row := db.DB.QueryRow("SELECT id, customer_email, total_amount, status, stripe_id, created_at, updated_at FROM orders WHERE stripe_id = ?", stripeID)

	order := &Order{}
	var statusStr string
	err := row.Scan(&order.ID, &order.CustomerEmail, &order.TotalAmount, &statusStr, &order.StripeID, &order.CreatedAt, &order.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order with Stripe ID %s not found", stripeID)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching order by Stripe ID %s: %w", stripeID, err)
	}

	order.Status = OrderStatus(statusStr)

	// Fetch order items
	rows, err := db.DB.Query("SELECT id, order_id, product_id, product_name, quantity, unit_price FROM order_items WHERE order_id = ?", order.ID)
	if err != nil {
		return order, fmt.Errorf("error fetching order items for order %s: %w", order.ID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPrice); err != nil {
			return order, fmt.Errorf("error scanning order item row for order %s: %w", order.ID, err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return order, fmt.Errorf("error after iterating through order item rows for order %s: %w", order.ID, err)
	}

	return order, nil
}

// UpdateOrderStatus updates the status of an order in the database.
func (o *Order) UpdateOrderStatus(status OrderStatus) error {
	_, err := db.DB.Exec("UPDATE orders SET status = ?, updated_at = ? WHERE id = ?", string(status), time.Now(), o.ID)
	if err != nil {
		return fmt.Errorf("error updating status for order %s: %w", o.ID, err)
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	log.Printf("Order %s status updated to %s.", o.ID, status)
	return nil
}
