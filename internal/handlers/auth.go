package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"threadly/internal/db"
	"threadly/internal/models"
	"threadly/internal/services"

	"github.com/gin-gonic/gin"
)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	// Remove non-alphanumeric except hyphens
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	slug = string(result)
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return ""
	}
	return slug
}

func uniqueSlug(base string) string {
	slug := base
	counter := 1
	for {
		var existing models.Business
		if db.DB.Where("slug = ?", slug).First(&existing).Error != nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, counter)
		counter++
	}
	return slug
}

func ShowLogin(c *gin.Context) {
	if token, err := c.Cookie("token"); err == nil && token != "" {
		if _, err := services.ValidateToken(token); err == nil {
			c.Redirect(http.StatusFound, "/business")
			return
		}
	}
	c.HTML(200, "business_login.html", gin.H{
		"Title": "Login - Threadly",
	})
}

func ShowRegister(c *gin.Context) {
	if token, err := c.Cookie("token"); err == nil && token != "" {
		if _, err := services.ValidateToken(token); err == nil {
			c.Redirect(http.StatusFound, "/business")
			return
		}
	}
	c.HTML(200, "register.html", gin.H{
		"Title": "Register - Threadly",
	})
}

func Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	name := c.PostForm("name")
	username := c.PostForm("username")
	businessType := c.PostForm("business_type")

	hashedPassword := services.Hash(password)

	slug := uniqueSlug(generateSlug(name))
	if slug == "" {
		slug = uniqueSlug(generateSlug(username))
	}

	user := models.Business{
		Email:        email,
		Password:     hashedPassword,
		Name:         name,
		Username:     username,
		BusinessType: businessType,
		Slug:         slug,
		IsPublic:     true,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.HTML(400, "register.html", gin.H{
			"Title": "Register - Threadly",
			"Error": "Email already exists",
		})
		return
	}

	c.Redirect(http.StatusFound, "/business/login")
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

	c.SetCookie("token", token, 86400, "/business", "", false, true)
	c.Redirect(http.StatusFound, "/business")
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/business", "", false, true)
	c.Redirect(http.StatusFound, "/business/login")
}
