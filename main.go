package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()
    
    // Load HTML templates
    r.LoadHTMLGlob("templates/*")
    
    // Serve static files (CSS, JS, images)
    r.Static("/static", "./static")
    
    // Route untuk halaman home
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "HRIS System",
            "message": "Selamat Datang di Sistem Database Karyawan",
        })
    })
    
    // Route untuk halaman employees
    r.GET("/employees", func(c *gin.Context) {
        employees := []map[string]interface{}{
            {"id": 1, "name": "John Doe", "department": "HC", "position": "Manager", "email": "john@company.com"},
            {"id": 2, "name": "Jane Smith", "department": "Finance", "position": "Staff", "email": "jane@company.com"},
            {"id": 3, "name": "Bob Johnson", "department": "IT", "position": "Developer", "email": "bob@company.com"},
        }
        
        c.HTML(http.StatusOK, "employees.html", gin.H{
            "title": "Data Karyawan",
            "employees": employees,
        })
    })
    
    // API endpoint (tetap ada untuk future use)
    r.GET("/api/employees", func(c *gin.Context) {
        employees := []map[string]interface{}{
            {"id": 1, "name": "John Doe", "department": "HC", "position": "Manager"},
            {"id": 2, "name": "Jane Smith", "department": "Finance", "position": "Staff"},
            {"id": 3, "name": "Bob Johnson", "department": "IT", "position": "Developer"},
        }
        c.JSON(http.StatusOK, gin.H{
            "data": employees,
            "total": len(employees),
        })
    })
    
    println("🚀 Server running on http://localhost:8080")
    r.Run(":8080")
}
