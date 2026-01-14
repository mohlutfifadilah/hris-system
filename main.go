package main

import (
	"fmt"
	"html/template"
	"log"

	"hris-system/config"
	migrations "hris-system/database/migration"
	seeders "hris-system/database/seeder"
	"hris-system/internal/controllers"
	"hris-system/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// ===== CONNECT DATABASE (akan di-skip jika USE_DATABASE=false) =====
	if err := config.ConnectDatabase(); err != nil {
		log.Println("⚠️  Running without database - some features may not work")
	}

	// ===== RUN MIGRATIONS & SEEDERS (hanya jika DB aktif) =====
	if config.DB != nil {
		log.Println("🔄 Running migrations...")
		if err := migrations.RunMigrations(); err != nil {
			log.Println("⚠️  Migration failed:", err)
		}

		log.Println("🌱 Running seeders...")
		if err := seeders.Seed(); err != nil {
			log.Println("⚠️  Seeding failed:", err)
		}
	} else {
		log.Println("⏭️  Skipping migrations & seeders (database disabled)")
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
	EmployeeController := controllers.NewEmployeeController()
	profileController := controllers.NewProfileController()
	departmentController := controllers.NewDepartmentController()
	gradingController := controllers.NewGradingController()
	statusController := controllers.NewStatusController()
	achievementController := controllers.NewAchievementController()
	typeAchievementController := controllers.NewTypeAchievement()

	// ===== PUBLIC ROUTES (tidak perlu login) =====
	r.GET("/", authController.ShowLoginForm)

	// Auth
	r.POST("/login", authController.Login)
	r.GET("/logout", authController.Logout)

	// Dashboard routes
	r.GET("/dashboard", dashboardController.Index)

	// Employee routes
	r.GET("/employee", EmployeeController.Index)
	r.GET("/employee/create", EmployeeController.Create)
	r.POST("/employee", EmployeeController.Store)
	r.GET("/employee/:id/edit", EmployeeController.Edit)
	r.POST("/employee/:id", EmployeeController.Update)
	r.POST("/employee/:id/delete", EmployeeController.Delete)

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
	r.GET("/grading/create", gradingController.Create)
	r.POST("/grading", gradingController.Store)
	r.GET("/grading/:id/edit", gradingController.Edit)
	r.POST("/grading/:id", gradingController.Update)
	r.POST("/grading/:id/delete", gradingController.Delete)

	// Status routes
	r.GET("/status", statusController.Index)
	r.GET("/status/create", statusController.Create)
	r.POST("/status", statusController.Store)
	r.GET("/status/:id/edit", statusController.Edit)
	r.POST("/status/:id", statusController.Update)
	r.POST("/status/:id/delete", statusController.Delete)

	// Achievement routes
	r.GET("/achievement", achievementController.Index)
	r.GET("/achievement/create", achievementController.Create)
	r.POST("/achievement", achievementController.Store)
	r.GET("/achievement/:id/show", achievementController.Show)
	r.GET("/achievement/generateExcel/:employee_id", achievementController.Excel)
	r.GET("/achievement/generatePdf/:employee_id", achievementController.Pdf)
	r.POST("/achievement/:id", achievementController.Update)
	r.POST("/achievement/:id/delete", achievementController.Delete)

	// Type Achievement routes
	r.GET("/typeAchievement", typeAchievementController.Index)
	r.GET("/typeAchievement/create", typeAchievementController.Create)
	r.POST("/typeAchievement", typeAchievementController.Store)
	r.GET("/typeAchievement/:id/edit", typeAchievementController.Edit)
	r.POST("/typeAchievement/:id", typeAchievementController.Update)
	r.POST("/typeAchievement/:id/delete", typeAchievementController.Delete)

	println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}

// loadTemplates - Load all templates dengan layout
func loadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"old": func(form map[string]string, key string) string {
			if form == nil {
				return ""
			}
			return form[key]
		},
		"err": func(errors map[string]string, key string) string {
			if errors == nil {
				return ""
			}
			return errors[key]
		},
		"selected": func(oldValue string, optionValue any) string {
			if oldValue == fmt.Sprint(optionValue) {
				return "selected"
			}
			return ""
		},
	}

	tmpl := template.New("").Funcs(funcMap)

	tmpl = template.Must(tmpl.ParseFiles(
		"templates/layouts/main.html",
		"templates/layouts/header.html",
		"templates/layouts/sidebar.html",
		"templates/layouts/footer.html",
	))
	tmpl = utils.RegisterTemplateHelpers(tmpl)

	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseFiles("templates/login.html"))

	return tmpl
}
