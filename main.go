package main

import (
	"html/template"

	"hris-system/config"
	"hris-system/internal/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

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
