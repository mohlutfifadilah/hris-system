package main

import (
    "net/http"
    "hris-system/config"
    "hris-system/models"
    "hris-system/internal/middleware"
    "hris-system/internal/controllers"
    
    "github.com/gin-gonic/gin"
)

func main() {
    // Connect to database
    config.ConnectDatabase()
    
    r := gin.Default()
    
    // Load HTML templates
    r.LoadHTMLGlob("templates/*")
    
    // Serve static files
    r.Static("/static", "./static")
    
    // Initialize controllers
    authController := controllers.NewAuthController()
    
    // Public routes (tidak perlu login)
    r.GET("/login", middleware.RedirectIfAuthenticated(), authController.ShowLoginForm)
    r.POST("/login", authController.Login)
    r.GET("/logout", authController.Logout)
    
    // Protected routes (harus login)
    protected := r.Group("/")
    protected.Use(middleware.AuthRequired())
    {
        protected.GET("/", homeHandler)
        protected.GET("/employees", employeesHandler)
        protected.GET("/api/employees", apiEmployeesHandler)
    }
    
    println("🚀 Server running on http://localhost:8080")
    r.Run(":8080")
}

func homeHandler(c *gin.Context) {
    userName := c.GetString("userName")
    
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":    "HRIS System",
        "message":  "Selamat Datang di Sistem Database Karyawan",
        "userName": userName,
    })
}

func employeesHandler(c *gin.Context) {
    var employees []models.Employee
    userName := c.GetString("userName")
    
    if err := config.DB.Find(&employees).Error; err != nil {
        c.HTML(http.StatusInternalServerError, "employees.html", gin.H{
            "title":     "Data Karyawan",
            "employees": []models.Employee{},
            "error":     "Gagal mengambil data",
            "userName":  userName,
        })
        return
    }
    
    c.HTML(http.StatusOK, "employees.html", gin.H{
        "title":     "Data Karyawan",
        "employees": employees,
        "userName":  userName,
    })
}

func apiEmployeesHandler(c *gin.Context) {
    var employees []models.Employee
    
    if err := config.DB.Find(&employees).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch data",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data":  employees,
        "total": len(employees),
    })
}
