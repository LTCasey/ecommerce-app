package handlers

import (
	"ecommerce-app/models"

	"github.com/gofiber/fiber/v2"
)

// RegisterProductRoutes registers all product-related routes
func RegisterProductRoutes(app *fiber.App) {
	app.Get("/products", ListProducts)
	app.Get("/products/:id", GetProduct)
}

// ListProducts renders the product listing page
func ListProducts(c *fiber.Ctx) error {
	products, err := models.GetProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load products")
	}
	return c.Render("products", fiber.Map{
		"Title":    "All Products",
		"Products": products,
	})
}

// GetProduct renders the product detail page
func GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	product, err := models.GetProductByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Redirect("/products")
	}

	return c.Render("product", fiber.Map{
		"Title":   product.Name,
		"Product": product,
	})
}
