package main

import (
	"html/template"
	"log"

	"hris-system/config"
	migrations "hris-system/database/migration"
	seeders "hris-system/database/seeder"
	"hris-system/internal/controllers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
	gradingController := controllers.NewGradingController()
	statusController := controllers.NewStatusController()
	typeAchievementController := controllers.NewTypeAchievement()

	// ===== PUBLIC ROUTES (tidak perlu login) =====
	r.GET("/", authController.ShowLoginForm)

	// Auth
	r.POST("/login", authController.Login)
	r.GET("/logout", authController.Logout)

	// Profile routes
	r.GET("/profile", profileController.Index)
	r.POST("/profile/upload-photo", profileController.UploadProfilePhoto)
	r.DELETE("/profile/delete-photo", profileController.DeleteProfilePhoto)

	// Department routes
	r.GET("/department", departmentController.Index)
	r.GET("/departments/create", departmentController.Create)
	r.POST("/departments", departmentController.Store)
	r.GET("/departments/:id/edit", departmentController.Edit)
	r.POST("/departments/:id", departmentController.Update)
	r.POST("/departments/:id/delete", departmentController.Delete)

	// Grading routes
	r.GET("/grading", gradingController.Index)
	r.GET("/gradings/create", gradingController.Create)
	r.POST("/gradings", gradingController.Store)
	r.GET("/gradings/:id/edit", gradingController.Edit)
	r.POST("/gradings/:id", gradingController.Update)
	r.POST("/gradings/:id/delete", gradingController.Delete)

	// Status routes
	r.GET("/status", statusController.Index)
	r.GET("/status/create", statusController.Create)
	r.POST("/status", statusController.Store)
	r.GET("/status/:id/edit", statusController.Edit)
	r.POST("/status/:id", statusController.Update)
	r.POST("/status/:id/delete", statusController.Delete)

	// Type Achievement routes
	r.GET("/typeAchievement", typeAchievementController.Index)
	r.GET("/typeAchievement/create", typeAchievementController.Create)
	r.POST("/typeAchievement", typeAchievementController.Store)
	r.GET("/typeAchievement/:id/edit", typeAchievementController.Edit)
	r.POST("/typeAchievement/:id", typeAchievementController.Update)
	r.POST("/typeAchievement/:id/delete", typeAchievementController.Delete)

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
