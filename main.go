package main

import (
	"html/template"

	"hris-system/config"
	"hris-system/internal/controllers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	// if err := migrations.RunMigrations(); err != nil {
	// 	log.Fatal("Failed to migrate:", err)
	// }

	// if err := seeders.Seed(); err != nil {
	// 	log.Fatal("Seeding failed:", err)
	// }

	r := gin.Default()

	store := cookie.NewStore([]byte("very-secret-key"))
	r.Use(sessions.Sessions("hris-session", store))

	// Load Templates
	r.SetHTMLTemplate(loadTemplates())

	// Serve static files
	r.Static("/static", "./static")

	// Initialize controllers
	authController := controllers.NewAuthController()
	dashboardController := controllers.NewDashboardController()
	profileController := controllers.NewProfileController()
	departmentController := controllers.NewDepartmentController()
	rankController := controllers.NewRankController()

	// ===== PUBLIC ROUTES (tidak perlu login) =====
	r.GET("/", authController.ShowLoginForm)

	// Auth
	r.POST("/login", authController.Login)
	r.GET("/logout", authController.Logout)

	// Dashboard routes
	r.GET("/profile", profileController.Index)

	// Department routes
	r.GET("/department", departmentController.Index)
	r.GET("/departments/create", departmentController.Create)
	r.POST("/departments", departmentController.Store)
	r.GET("/departments/:id/edit", departmentController.Edit)
	r.POST("/departments/:id", departmentController.Update)
	r.POST("/departments/:id/delete", departmentController.Delete)

	// Rank routes
	r.GET("/rank", rankController.Index)
	r.GET("/ranks/create", rankController.Create)
	r.POST("/ranks", rankController.Store)
	r.GET("/ranks/:id/edit", rankController.Edit)
	r.POST("/ranks/:id", rankController.Update)
	r.POST("/ranks/:id/delete", rankController.Delete)

	// Dashboard routes
	r.GET("/dashboard", dashboardController.Index)

	println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}

// loadTemplates - Load all templates dengan layout
func loadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	// Mulai dari template kosong + funcMap
	tmpl := template.New("").Funcs(funcMap)

	tmpl = template.Must(tmpl.ParseFiles(
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
