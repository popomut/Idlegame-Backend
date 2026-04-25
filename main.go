package main

import (
        "fmt"
        "log"

        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/fiber/v2/middleware/cors"

        "idlegame-backend/database"
        "idlegame-backend/handlers"
        "idlegame-backend/middleware"
)

func main() {
        // Initialize database
        err := database.Init()
        if err != nil {
                log.Fatalf("Failed to initialize database: %v", err)
        }
        defer database.Close()

        // Create Fiber app
        app := fiber.New(fiber.Config{
                Prefork: false,
        })

        // Middleware
        app.Use(cors.New(cors.Config{
                AllowOriginsFunc: func(origin string) bool { return true },
                AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
                AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
                AllowCredentials: true,
        }))

        // Public routes (no auth required)
        app.Post("/api/auth/register", handlers.Register)
        app.Post("/api/auth/login", handlers.Login)
        app.Post("/api/auth/guest", handlers.GuestLogin)
        app.Post("/api/auth/logout", handlers.Logout)

        // Protected routes (require JWT token)
        api := app.Group("/api", middleware.AuthMiddleware())

        // User routes
        api.Get("/user", handlers.GetUser)
        api.Post("/user/update", handlers.UpdateUser)

        // Mining routes
        api.Post("/mining/start", handlers.StartMining)
        api.Post("/mining/stop", handlers.StopMining)
        api.Get("/mining/status", handlers.GetMiningStatus)

        // Inventory routes
        api.Get("/inventory/ores", handlers.GetOreInventory)

        // Health check
        app.Get("/health", func(c *fiber.Ctx) error {
                return c.JSON(fiber.Map{"status": "ok"})
        })

        // Start server
        port := 5000
        fmt.Printf("🚀 Server running on http://0.0.0.0:%d\n", port)
        log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%d", port)))
}
