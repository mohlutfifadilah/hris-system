package main

import (
	"html/template"
	"log"

	"hris-system/config"
	migrations "hris-system/database/migration"
	seeders "hris-system/database/seeder"
	"hris-system/internal/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	if err := migrations.RunMigrations(); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	if err := seeders.Seed(); err != nil {
		log.Fatal("Seeding failed:", err)
	}

	r := gin.Default()

	// Load Templates
	r.SetHTMLTemplate(loadTemplates())

	// Serve static files
	r.Static("/static", "./static")

	// Initialize controllers
	authController := controllers.NewAuthController()
	dashboardController := controllers.NewDashboardController()

	// ===== PUBLIC ROUTES (tidak perlu login) =====
	r.GET("/", authController.ShowLoginForm)

	// Auth
	r.POST("/login", authController.Login)
	r.GET("/logout", authController.Logout)

	// Dashboard routes
	r.GET("/dashboard", dashboardController.Index)

	println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}

// loadTemplates - Load all templates dengan layout
func loadTemplates() *template.Template {
	tmpl := template.Must(template.ParseFiles(
		"templates/layouts/main.html",
		"templates/layouts/header.html",
		"templates/layouts/sidebar.html",
		"templates/layouts/footer.html",
	))

	// Parse dashboard templates
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

	// Parse login template
	tmpl = template.Must(tmpl.ParseFiles("templates/login.html"))

	return tmpl
}
