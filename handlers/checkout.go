package handlers

import (
	"ecommerce-app/models"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Cart session key
const SessionCartKey = "cart"

var sessionStore *session.Store // Declare sessionStore here

func InitSessionStore(s *session.Store) {
	sessionStore = s
}

// RegisterCheckoutRoutes registers all checkout-related routes
func RegisterCheckoutRoutes(app *fiber.App) {
	app.Get("/cart", ViewCart)
	app.Post("/cart/add/:id", AddToCart)
	app.Post("/cart/remove/:id", RemoveFromCart)
	app.Get("/checkout", Checkout)
	app.Post("/checkout", Checkout)
	app.Get("/checkout/success", CheckoutSuccess)
	app.Get("/checkout/cancel", CheckoutCancel)
	app.Post("/webhook/stripe", StripeWebhook)
}

// ViewCart displays the current shopping cart
func ViewCart(c *fiber.Ctx) error {
	cart := getCart(c)

	var totalAmount float64
	for _, item := range cart.Items {
		totalAmount += item.UnitPrice * float64(item.Quantity)
	}

	return c.Render("cart", fiber.Map{
		"Title":    "Your Shopping Cart",
		"Cart":     cart,
		"Total":    fmt.Sprintf("$%.2f", totalAmount),
		"HasItems": len(cart.Items) > 0,
	})
}

// AddToCart adds a product to the cart
func AddToCart(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Get quantity from form, default to 1
	quantity := 1
	if c.FormValue("quantity") != "" {
		var err error
		quantity, err = strconv.Atoi(c.FormValue("quantity"))
		if err != nil || quantity < 1 {
			quantity = 1
		}
	}

	// Get product
	product, err := models.GetProductByID(productID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Redirect("/products")
	}

	// Get the current cart from session
	cart := getCart(c)

	// Add product to cart
	cart.AddItem(product, quantity)

	// Save cart to session
	saveCart(c, cart)

	// Redirect back to products or to cart
	return c.Redirect("/cart")
}

// RemoveFromCart removes a product from the cart
func RemoveFromCart(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Get the current cart from session
	cart := getCart(c)

	// Remove item from cart
	for i, item := range cart.Items {
		if item.ProductID == productID {
			// Remove the item from the slice
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			break
		}
	}

	// Recalculate total
	cart.CalculateTotal()

	// Save cart to session
	saveCart(c, cart)

	return c.Redirect("/cart")
}

// Checkout processes the checkout and redirects to Stripe
func Checkout(c *fiber.Ctx) error {
	// Get the current cart
	cart := getCart(c)

	// Make sure we have items in cart
	if len(cart.Items) == 0 {
		return c.Redirect("/cart")
	}

	// Get email from form
	email := c.FormValue("email")
	if email == "" {
		return c.Render("checkout_email", fiber.Map{
			"Title": "Enter Email to Checkout",
		})
	}

	// Create an order from the cart
	order := models.NewOrder(email)

	// Copy items from cart to order
	for _, item := range cart.Items {
		product, _ := models.GetProductByID(item.ProductID)
		order.AddItem(product, item.Quantity)
	}

	// Save the order to the database *before* creating the Stripe session
	// This ensures the order exists when the webhook is received.
	err := order.Save()
	if err != nil {
		log.Printf("Error saving order to database before checkout: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error processing order")
	}

	// Create success and cancel URLs
	successURL := fmt.Sprintf("%s/checkout/success?session_id={CHECKOUT_SESSION_ID}", c.BaseURL())
	cancelURL := fmt.Sprintf("%s/checkout/cancel", c.BaseURL())

	// Create Stripe checkout session
	checkoutURL, err := models.CreateCheckoutSession(order, successURL, cancelURL)
	if err != nil {
		log.Printf("Error creating Stripe checkout session: %v", err)
		// Depending on your error handling, you might want to mark the order as failed
		// order.UpdateOrderStatus(models.OrderStatusFailed)
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating checkout session")
	}

	// The Stripe Session ID is set within models.CreateCheckoutSession

	// Clear the cart after creating the order and checkout session
	clearCart(c)

	// Redirect to Stripe checkout URL
	return c.Redirect(checkoutURL, fiber.StatusSeeOther)
}

// CheckoutSuccess handles successful checkout
func CheckoutSuccess(c *fiber.Ctx) error {
	// The Stripe webhook handler will update the order status in the DB

	// Get the session ID from the query parameter
	sessionID := c.Query("session_id")
	if sessionID == "" {
		// Handle case where session_id is missing
		log.Println("Checkout success called without session_id")
		return c.Render("checkout_success", fiber.Map{
			"Title":   "Order Complete",
			"Message": "Your order is complete!",
		})
	}

	// Retrieve the order from the database using the sessionID
	order, err := models.GetOrderByStripeID(sessionID)
	if err != nil {
		log.Printf("Error retrieving order for success page (Stripe ID %s): %v", sessionID, err)
		// Continue rendering success page even if order retrieval fails
		return c.Render("checkout_success", fiber.Map{
			"Title":   "Order Complete",
			"Message": "Your order is complete! We could not retrieve order details at this moment.",
		})
	}

	// Render the success page with order details
	return c.Render("checkout_success", fiber.Map{
		"Title":   "Order Complete",
		"Message": fmt.Sprintf("Thank you for your order, %s! Your order ID is %s.", order.CustomerEmail, order.ID),
		"Order":   order,
	})
}

// CheckoutCancel handles cancelled checkout
func CheckoutCancel(c *fiber.Ctx) error {
	// Optionally retrieve the order based on session_id (if passed) and update status to cancelled
	// For this example, we just render the cancel page.

	// Get the session ID from the query parameter (optional)
	sessionID := c.Query("session_id")
	if sessionID != "" {
		// Attempt to retrieve the order and update its status to cancelled
		order, err := models.GetOrderByStripeID(sessionID)
		if err != nil {
			log.Printf("Error retrieving order for cancel page (Stripe ID %s): %v", sessionID, err)
		} else {
			err = order.UpdateOrderStatus(models.OrderStatusFailed) // Assuming 'failed' or similar indicates cancellation
			if err != nil {
				log.Printf("Error updating order status to cancelled for order %s: %v", order.ID, err)
			}
		}
	}

	return c.Render("checkout_cancel", fiber.Map{
		"Title":   "Checkout Cancelled",
		"Message": "Your checkout was cancelled. You can continue shopping.",
	})
}

// StripeWebhook handles Stripe webhook events
func StripeWebhook(c *fiber.Ctx) error {
	// This handler will be fully implemented in the next step.

	// Get the signature header
	signature := c.Get("Stripe-Signature")

	// Read the request body
	payload := c.Body()

	// Process the webhook using the function in models/stripe.go
	err := models.HandleStripeWebhook(payload, signature)
	if err != nil {
		log.Printf("Stripe webhook handling error: %v", err)
		return c.SendStatus(fiber.StatusBadRequest) // Return 400 for invalid signature or processing errors
	}

	// Return 200 OK for successful processing
	return c.SendStatus(fiber.StatusOK)
}

// Helper function to get cart from session
func getCart(c *fiber.Ctx) *models.Order {
	// Get the session store
	sess, err := sessionStore.Get(c)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		// Return a new empty cart in case of session error
		return models.NewOrder("")
	}

	// Get cart from session
	cartData := sess.Get(SessionCartKey)
	if cartData == nil {
		// Create a new cart if none exists
		cart := models.NewOrder("")
		saveCart(c, cart)
		return cart
	}

	// Type assert cart data
	cart, ok := cartData.(*models.Order)
	if !ok {
		log.Println("Error asserting cart data from session, creating new cart.")
		// Return a new empty cart if type assertion fails
		cart := models.NewOrder("")
		saveCart(c, cart)
		return cart
	}

	return cart
}

// Helper function to save cart to session
func saveCart(c *fiber.Ctx, cart *models.Order) {
	// Get the session store
	sess, err := sessionStore.Get(c)
	if err != nil {
		log.Printf("Error getting session to save cart: %v", err)
		return // Cannot save cart if session store is unavailable
	}

	// Set cart in session
	sess.Set(SessionCartKey, cart)

	// Save session
	if err := sess.Save(); err != nil {
		log.Printf("Error saving session with cart: %v", err)
	}
}

// Helper function to clear the cart from session
func clearCart(c *fiber.Ctx) {
	// Get the session store
	sess, err := sessionStore.Get(c)
	if err != nil {
		log.Printf("Error getting session to clear cart: %v", err)
		return // Cannot clear cart if session store is unavailable
	}

	// Delete cart from session
	sess.Delete(SessionCartKey)

	// Save session
	if err := sess.Save(); err != nil {
		log.Printf("Error saving session after clearing cart: %v", err)
	}
}
