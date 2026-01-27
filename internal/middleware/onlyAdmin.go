package middleware

import (
	"net/http"

	auth "hris-system/internal/auth"

	"github.com/gin-gonic/gin"
)

// AdminOnlyMiddleware - Middleware untuk memastikan hanya admin yang bisa mengakses route tertentu
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil user yang sedang login dari context
		currentUser := auth.GetCurrentUser(c) // Mengambil user dari session atau token
		if currentUser == nil {
			c.SetCookie("flash_message", "Please Login First", 5, "/", "", false, false)
			c.SetCookie("token", "", -1, "/", "", false, true) // Hapus token
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}

		// Periksa apakah user adalah admin
		if currentUser.IsAdmin == false {
			c.SetCookie("flash_message", "You do not have permission to access this route.", 5, "/", "", false, false)
			// Mendapatkan referer (halaman sebelumnya)
			referer := c.Request.Referer()
			if referer == "" {
				// Jika referer kosong, arahkan ke halaman home atau halaman lain yang diinginkan
				referer = "/"
			}
			// Redirect ke halaman sebelumnya (referer)
			c.Redirect(http.StatusSeeOther, referer)
			c.Abort()
			return
		}

		// Lanjutkan ke handler berikutnya jika user adalah admin
		c.Next()
	}
}
