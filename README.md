# E-Commerce App with Fiber and Stripe

A complete e-commerce web application built with Go Fiber framework and Stripe payment integration.

> **Security Notice:**
> - **Never commit your real API keys, secrets, or sensitive data to this repository.**
> - Use environment variables or a `.env` file (which should be added to `.gitignore`) to manage secrets locally.
> - If sharing this project, provide a `.env.example` file with only variable names and placeholder values.
> - **Never share your Stripe secret keys publicly.**

## Features

- Product listing and detail pages
- Shopping cart functionality
- Checkout process with Stripe integration
- Responsive design with Bootstrap

## Prerequisites

- Go 1.20 or higher
- Stripe account for payment processing

## Installation

1. Clone the repository:
```
git clone https://github.com/LTCasey/ecommerce-app.git
cd ecommerce-app
```

2. Install dependencies:
```
go mod download
```

3. Set up environment variables:
```
# Linux/Mac
export STRIPE_SECRET_KEY=your_stripe_secret_key
export STRIPE_WEBHOOK_SECRET=your_webhook_secret_key

# Windows
set STRIPE_SECRET_KEY=your_stripe_secret_key
set STRIPE_WEBHOOK_SECRET=your_webhook_secret_key
```

Alternatively, create a `.env` file in your project root (do not commit this file):
```
STRIPE_SECRET_KEY=your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=your_webhook_secret_key
```

> **Tip:** Add `.env` to your `.gitignore` to prevent accidental commits of sensitive data.

## Running the Application

1. Ensure your environment variables are set as described above.
2. Start the server:
```
go run main.go
```
3. Open your browser and navigate to `http://localhost:3000` to use the app.

## How the App Runs

- **Browse Products:** Users can view a list of products and see details for each product.
- **Add to Cart:** Users can add products to their shopping cart.
- **View Cart:** The cart page shows all selected items and the total price.
- **Checkout:** Users proceed to checkout, where payment is handled securely via Stripe Checkout.
- **Order Completion:** After successful payment, users receive confirmation, and the order is processed.