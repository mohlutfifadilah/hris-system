package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"hris-system/config"
	migrations "hris-system/database/migration"
	seeders "hris-system/database/seeder"
	"hris-system/internal/controllers"
	middleware "hris-system/internal/middleware"
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

	// Contoh di main.go
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Initialize controllers
	authController := controllers.NewAuthController()
	dashboardController := controllers.NewDashboardController()
	EmployeeController := controllers.NewEmployeeController()
	profileController := controllers.NewProfileController()
	departmentController := controllers.NewDepartmentController()
	careerController := controllers.NewCareerController()
	gradingController := controllers.NewGradingController()
	statusController := controllers.NewStatusController()
	achievementController := controllers.NewAchievementController()
	typeAchievementController := controllers.NewTypeAchievement()

	// ===== PUBLIC ROUTES (tidak perlu login) =====
	r.GET("/", authController.ShowLoginForm)

	// Auth
	r.POST("/login", authController.Login)
	r.GET("/logout", authController.Logout)

	// Employee routes
	r.GET("/employee", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Index)
	r.GET("/employee/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Create)
	r.POST("/employee", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Store)
	r.GET("/employee/:id/edit", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Edit)
	r.POST("/employee/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Update)
	r.POST("/employee/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), EmployeeController.Delete)

	// Profile routes
	r.POST("/profile/upload-photo", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), profileController.UploadProfilePhoto)
	r.DELETE("/profile/delete-photo", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), profileController.DeleteProfilePhoto)

	// Career routes
	r.GET("/career", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), careerController.Index)
	r.GET("/career/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), careerController.Create)
	r.POST("/career", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), careerController.Store)
	r.POST("/career/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), careerController.Update)
	r.POST("/career/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), careerController.Delete)

	// Department routes
	r.GET("/department", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Index)
	r.GET("/departments/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Create)
	r.POST("/departments", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Store)
	r.GET("/departments/:id/edit", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Edit)
	r.POST("/departments/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Update)
	r.POST("/departments/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), departmentController.Delete)

	// Grading routes
	r.GET("/grading", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Index)
	r.GET("/grading/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Create)
	r.POST("/grading", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Store)
	r.GET("/grading/:id/edit", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Edit)
	r.POST("/grading/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Update)
	r.POST("/grading/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), gradingController.Delete)

	// Status routes
	r.GET("/status", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Index)
	r.GET("/status/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Create)
	r.POST("/status", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Store)
	r.GET("/status/:id/edit", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Edit)
	r.POST("/status/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Update)
	r.POST("/status/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), statusController.Delete)

	// Achievement routes
	r.GET("/achievement", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), achievementController.Index)
	r.GET("/achievement/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), achievementController.Create)
	r.POST("/achievement", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), achievementController.Store)
	r.POST("/achievement/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), achievementController.Update)
	r.POST("/achievement/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), achievementController.Delete)

	// Type Achievement routes
	r.GET("/typeAchievement", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Index)
	r.GET("/typeAchievement/create", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Create)
	r.POST("/typeAchievement", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Store)
	r.GET("/typeAchievement/:id/edit", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Edit)
	r.POST("/typeAchievement/:id", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Update)
	r.POST("/typeAchievement/:id/delete", middleware.AuthRequired(), middleware.AdminOnlyMiddleware(), typeAchievementController.Delete)

	// Dashboard routes
	r.GET("/dashboard", middleware.AuthRequired(), dashboardController.Index)

	// Employee routes
	r.GET("/employee/:id/show", middleware.AuthRequired(), EmployeeController.Show)

	// Profile routes
	r.GET("/profile", middleware.AuthRequired(), profileController.Index)
	r.POST("/profile/change-password", middleware.AuthRequired(), profileController.ChangePassword)

	// Career routes
	r.GET("/career/:id/show", middleware.AuthRequired(), careerController.Show)
	r.GET("/career/generateExcel/:employee_id", middleware.AuthRequired(), careerController.Excel)
	r.GET("/career/generatePdf/:employee_id", middleware.AuthRequired(), careerController.Pdf)

	// Achievement routes
	r.GET("/achievement/:id/show", middleware.AuthRequired(), achievementController.Show)
	r.GET("/achievement/generateExcel/:employee_id", middleware.AuthRequired(), achievementController.Excel)
	r.GET("/achievement/generatePdf/:employee_id", middleware.AuthRequired(), achievementController.Pdf)

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
