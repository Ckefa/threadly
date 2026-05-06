package handlers

import (
	"net/http"

	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/services"

	"github.com/gin-gonic/gin"
)

func ShowLogin(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{
		"Title": "Login - Threadly",
	})
}

func ShowRegister(c *gin.Context) {
	c.HTML(200, "register.html", gin.H{
		"Title": "Register - Threadly",
	})
}

func Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	businessType := c.PostForm("business_type")

	hashedPassword := services.Hash(password)

	user := models.Business{
		Email:        email,
		Password:     hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		BusinessType: businessType,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.HTML(400, "register.html", gin.H{
			"Title": "Register - Threadly",
			"Error": "Email already exists",
		})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	var user models.Business
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(401, "login.html", gin.H{
			"Title": "Login - Threadly",
			"Error": "Invalid email or password",
		})
		return
	}

	if !services.Check(user.Password, password) {
		c.HTML(401, "login.html", gin.H{
			"Title": "Login - Threadly",
			"Error": "Invalid email or password",
		})
		return
	}

	token, err := services.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.HTML(500, "login.html", gin.H{
			"Title": "Login - Threadly",
			"Error": "Failed to generate token",
		})
		return
	}

	c.SetCookie("token", token, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
