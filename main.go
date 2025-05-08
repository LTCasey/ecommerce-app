package main

import (
	"ecommerce-app/db"
	"ecommerce-app/handlers"
	"ecommerce-app/models"
	"encoding/gob"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file")
	}
	// Register types for session encoding/decoding
	gob.Register(&models.Order{})
}

func main() {
	// Initialize Stripe
	models.InitStripe()

	// Initialize Database
	db.InitDB()

	// Seed Products into the database
	err := models.SeedProducts()
	if err != nil {
		log.Fatalf("Error seeding products: %v", err)
	}

	// Initialize HTML Templates
	engine := html.New("./views", ".html")

	// Set up the Fiber app with the template engine
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layout", // Use layout.html as the base template
	})

	// Initialize Session Store
	store := session.New(session.Config{
		Expiration: time.Hour * 24,
		CookiePath: "/",
		// You can configure other options like KeyGenerator, Storage, etc.
	})

	// Pass the session store to handlers that need it (like checkout)
	handlers.InitSessionStore(store)

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Static files
	app.Static("/static", "./static")

	// Setup routes
	setupRoutes(app)

	// Get port from environment variables or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start the server
	log.Printf("Server starting on port %s...", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	// Home page
	app.Get("/", func(c *fiber.Ctx) error {
		products, err := models.GetProducts()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error fetching products")
		}
		// Show only first 3 products on home page, if available
		displayProducts := products
		if len(products) > 3 {
			displayProducts = products[:3]
		}
		return c.Render("index", fiber.Map{
			"Title":    "Welcome",
			"Products": displayProducts,
		})
	})

	// Register product routes (listing, details)
	handlers.RegisterProductRoutes(app)

	// Register checkout routes (cart, checkout, payment)
	handlers.RegisterCheckoutRoutes(app)
}
