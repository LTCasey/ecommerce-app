# E-Commerce App with Fiber and Stripe

A complete e-commerce web application built with Go Fiber framework and Stripe payment integration.

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
git clone https://github.com/yourusername/ecommerce-app.git
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

## Running the Application

Start the server:
```
go run main.go
```

The application will be available at `http://localhost:3000`

## Project Structure

- `/handlers` - HTTP request handlers
- `/models` - Data models and business logic
- `/static` - Static assets (CSS, JavaScript, images)
- `/views` - HTML templates
- `main.go` - Application entry point

## Stripe Integration

This application uses Stripe Checkout for payment processing. To fully use this feature:

1. Create a Stripe account at https://stripe.com
2. Get your API keys from the Stripe dashboard
3. Set the environment variables as described in the Installation section
4. For webhook testing, use Stripe CLI or a service like ngrok to forward webhook events to your local server

## Production Deployment

For production deployment, consider the following steps:

1. Set up a proper database for product and order storage
2. Configure proper error handling and logging
3. Set up TLS/SSL for secure communication
4. Configure proper session management
5. Implement user authentication 