package models

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/stripe/stripe-go/v74"
	checkoutsession "github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Initialize Stripe with your API key
func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		fmt.Println("Warning: STRIPE_SECRET_KEY environment variable not set")
		// For development purposes you might set a default key here
		// stripe.Key = "sk_test_..."
	}
}

// CreateCheckoutSession creates a new Stripe checkout session for the order
func CreateCheckoutSession(order *Order, successURL, cancelURL string) (string, error) {
	// Create line items from order items
	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, item := range order.Items {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        stripe.String(item.ProductName),
					Description: stripe.String(fmt.Sprintf("Order Item: %s", item.ProductID)),
				},
				UnitAmount: stripe.Int64(int64(item.UnitPrice * 100)), // Convert to cents
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	// Create checkout session parameters
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// CustomerEmail: stripe.String(order.CustomerEmail), // TODO: Add customer email to order or retrieve from user session
	}

	// Create the checkout session
	s, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Update order with Stripe session ID *after* successful session creation
	order.StripeID = s.ID
	// Save the order with the Stripe ID. This is crucial.
	err = order.Save()
	if err != nil {
		// Log error but continue, as the Stripe session is created
		log.Printf("Warning: Could not update order %s with Stripe ID %s: %v", order.ID, s.ID, err)
		// Depending on requirements, you might want to cancel the Stripe session here.
	}

	return s.URL, nil
}

// HandleStripeWebhook processes Stripe webhook events
func HandleStripeWebhook(payload []byte, signature string) error {
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET environment variable not set")
	}

	event, err := webhook.ConstructEvent(payload, signature, endpointSecret)
	if err != nil {
		return fmt.Errorf("error verifying webhook signature: %w", err)
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		// Payment is successful, update order status
		var stripeSessionData stripe.CheckoutSession              // Use the correct type
		err := json.Unmarshal(event.Data.Raw, &stripeSessionData) // Use json.Unmarshal
		if err != nil {
			return fmt.Errorf("error parsing webhook JSON for checkout.session.completed: %w", err)
		}

		log.Printf("Checkout session completed: %s", stripeSessionData.ID)

		// Retrieve the order from your database using the Stripe Session ID
		order, err := GetOrderByStripeID(stripeSessionData.ID)
		if err != nil {
			// This is a critical error: received webhook for unknown order
			return fmt.Errorf("order with Stripe ID %s not found in DB: %w", stripeSessionData.ID, err)
		}

		// Update the order status to completed
		if order.Status != OrderStatusCompleted {
			err = order.UpdateOrderStatus(OrderStatusCompleted)
			if err != nil {
				return fmt.Errorf("error updating order %s status to completed: %w", order.ID, err)
			}
			log.Printf("Order %s status updated to completed.", order.ID)

			// TODO: Implement post-checkout actions here (e.g., send order confirmation email)
		}

	case "checkout.session.expired":
		// Payment expired, update order status
		var stripeSessionData stripe.CheckoutSession              // Use the correct type
		err := json.Unmarshal(event.Data.Raw, &stripeSessionData) // Use json.Unmarshal
		if err != nil {
			return fmt.Errorf("error parsing webhook JSON for checkout.session.expired: %w", err)
		}

		log.Printf("Checkout session expired: %s", stripeSessionData.ID)

		// Retrieve the order from your database using the Stripe Session ID
		order, err := GetOrderByStripeID(stripeSessionData.ID)
		if err != nil {
			log.Printf("Warning: Order with Stripe ID %s not found in DB on expiry webhook: %v", stripeSessionData.ID, err)
			return nil // Not a critical error if order isn't found on expiry
		}

		// Update the order status to failed (or expired)
		if order.Status == OrderStatusPending {
			err = order.UpdateOrderStatus(OrderStatusFailed)
			if err != nil {
				return fmt.Errorf("error updating order %s status to failed on expiry: %w", order.ID, err)
			}
			log.Printf("Order %s status updated to failed (expired).", order.ID)
		}

	default:
		log.Printf("Unhandled Stripe event type: %s", event.Type)
	}

	return nil
}
