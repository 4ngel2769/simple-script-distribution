package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Admin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password_hash"`
	} `yaml:"admin"`
	Scripts []ScriptConfig `yaml:"scripts"`
}

type ScriptConfig struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
	Type        string `yaml:"type"` // "local" or "redirect"
	RedirectURL string `yaml:"redirect_url,omitempty"`
}

type IndexPageData struct {
	Scripts []ScriptConfig `json:"scripts"`
}

var (
	config       Config
	scriptsPath  string
	store        *session.Store
)

func main() {
	// Initialize
	loadConfig()
	scriptsPath = os.Getenv("SCRIPTS_PATH")
	if scriptsPath == "" {
		scriptsPath = "/app/scripts"
	}

	// Initialize session store
	store = session.New()

	// Initialize template engine
	engine := html.New("./templates", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Static files
	app.Static("/static", "./static")

	// Routes
	app.Get("/", indexHandler)
	app.Post("/login", loginHandler)
	app.Get("/admin", authMiddleware, adminHandler)
	app.Get("/admin/scripts", authMiddleware, getScriptsAPI)
	app.Post("/admin/scripts", authMiddleware, createScriptAPI)
	app.Put("/admin/scripts/:name", authMiddleware, updateScriptAPI)
	app.Delete("/admin/scripts/:name", authMiddleware, deleteScriptAPI)
	app.Get("/admin/scripts/:name/content", authMiddleware, getScriptContentAPI)
	app.Put("/admin/scripts/:name/content", authMiddleware, updateScriptContentAPI)
	app.Post("/admin/index-page", authMiddleware, updateIndexPageAPI)
	app.Get("/admin/index-page", authMiddleware, getIndexPageAPI)
	app.Post("/logout", logoutHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Admin dashboard starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func loadConfig() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("Failed to read config file:", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal("Failed to parse config file:", err)
	}
}

func saveConfig() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func indexHandler(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"Title": "Script Server Admin",
	})
}

func loginHandler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == config.Admin.Username {
		if err := bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(password)); err == nil {
			sess, _ := store.Get(c)
			sess.Set("authenticated", true)
			sess.Set("username", username)
			sess.Save()
			return c.Redirect("/admin")
		}
	}

	return c.Render("login", fiber.Map{
		"Title": "Script Server Admin",
		"Error": "Invalid credentials",
	})
}

func logoutHandler(c *fiber.Ctx) error {
	sess, _ := store.Get(c)
	sess.Destroy()
	return c.Redirect("/")
}

func authMiddleware(c *fiber.Ctx) error {
	sess, _ := store.Get(c)
	
	if auth := sess.Get("authenticated"); auth != true {
		return c.Redirect("/")
	}
	
	return c.Next()
}

func adminHandler(c *fiber.Ctx) error {
	return c.Render("admin", fiber.Map{
		"Title":   "Admin Dashboard",
		"Scripts": config.Scripts,
	})
}

func getScriptsAPI(c *fiber.Ctx) error {
	return c.JSON(config.Scripts)
}

func createScriptAPI(c *fiber.Ctx) error {
	var script ScriptConfig
	if err := c.BodyParser(&script); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if script.Name == "" || script.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and description are required"})
	}

	// Check if script already exists
	for _, existing := range config.Scripts {
		if existing.Name == script.Name {
			return c.Status(409).JSON(fiber.Map{"error": "Script already exists"})
		}
	}

	// Set defaults
	if script.Type == "" {
		script.Type = "local"
	}
	if script.Icon == "" {
		script.Icon = "📜"
	}
	if script.Path == "" {
		script.Path = script.Name
	}

	config.Scripts = append(config.Scripts, script)
	
	if err := saveConfig(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
	}

	// Create script directory and file if local type
	if script.Type == "local" {
		scriptDir := filepath.Join(scriptsPath, script.Name)
		os.MkdirAll(scriptDir, 0755)
		
		scriptFile := filepath.Join(scriptDir, fmt.Sprintf("runme_%s.sh", script.Name))
		defaultContent := fmt.Sprintf("#!/bin/bash\n\n# %s\n# Generated on %s\n\necho \"Hello from %s script!\"\n", 
			script.Description, time.Now().Format("2006-01-02 15:04:05"), script.Name)
		
		os.WriteFile(scriptFile, []byte(defaultContent), 0755)
	}

	return c.JSON(script)
}

func updateScriptAPI(c *fiber.Ctx) error {
	name := c.Params("name")
	var updates ScriptConfig
	
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	for i, script := range config.Scripts {
		if script.Name == name {
			// Update fields
			if updates.Description != "" {
				config.Scripts[i].Description = updates.Description
			}
			if updates.Icon != "" {
				config.Scripts[i].Icon = updates.Icon
			}
			if updates.Type != "" {
				config.Scripts[i].Type = updates.Type
			}
			if updates.RedirectURL != "" {
				config.Scripts[i].RedirectURL = updates.RedirectURL
			}

			if err := saveConfig(); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
			}

			return c.JSON(config.Scripts[i])
		}
	}

	return c.Status(404).JSON(fiber.Map{"error": "Script not found"})
}

func deleteScriptAPI(c *fiber.Ctx) error {
	name := c.Params("name")

	for i, script := range config.Scripts {
		if script.Name == name {
			// Remove from config
			config.Scripts = append(config.Scripts[:i], config.Scripts[i+1:]...)
			
			if err := saveConfig(); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
			}

			// Remove script directory if local type
			if script.Type == "local" {
				scriptDir := filepath.Join(scriptsPath, script.Name)
				os.RemoveAll(scriptDir)
			}

			return c.JSON(fiber.Map{"message": "Script deleted successfully"})
		}
	}

	return c.Status(404).JSON(fiber.Map{"error": "Script not found"})
}

func getScriptContentAPI(c *fiber.Ctx) error {
	name := c.Params("name")

	for _, script := range config.Scripts {
		if script.Name == name && script.Type == "local" {
			scriptFile := filepath.Join(scriptsPath, script.Name, fmt.Sprintf("runme_%s.sh", script.Name))
			content, err := os.ReadFile(scriptFile)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "Script file not found"})
			}

			return c.JSON(fiber.Map{"content": string(content)})
		}
	}

	return c.Status(404).JSON(fiber.Map{"error": "Script not found or not local"})
}

func updateScriptContentAPI(c *fiber.Ctx) error {
	name := c.Params("name")
	
	var req struct {
		Content string `json:"content"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	for _, script := range config.Scripts {
		if script.Name == name && script.Type == "local" {
			scriptFile := filepath.Join(scriptsPath, script.Name, fmt.Sprintf("runme_%s.sh", script.Name))
			
			if err := os.WriteFile(scriptFile, []byte(req.Content), 0755); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save script"})
			}

			return c.JSON(fiber.Map{"message": "Script updated successfully"})
		}
	}

	return c.Status(404).JSON(fiber.Map{"error": "Script not found or not local"})
}

func updateIndexPageAPI(c *fiber.Ctx) error {
	var data IndexPageData
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Generate new index.html
	htmlContent := generateIndexHTML(data.Scripts)
	
	indexPath := filepath.Join(scriptsPath, "index.html")
	if err := os.WriteFile(indexPath, []byte(htmlContent), 0644); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update index page"})
	}

	return c.JSON(fiber.Map{"message": "Index page updated successfully"})
}

func getIndexPageAPI(c *fiber.Ctx) error {
	return c.JSON(IndexPageData{Scripts: config.Scripts})
}

func generateIndexHTML(scripts []ScriptConfig) string {
	var scriptElements strings.Builder
	
	for _, script := range scripts {
		scriptElements.WriteString(fmt.Sprintf(`        <div class="endpoint" data-script="%s">
            <span class="emoji">%s</span>/%s - %s
            <div class="copy-feedback">Copied!</div>
        </div>
        
`, script.Name, script.Icon, script.Name, script.Description))
	}

	// Return the full HTML with updated script elements
	// This is a simplified version - in practice you might want to use templates
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Script Server</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!-- Include the full CSS and JS from the original file -->
</head>
<body>
    <div class="container">
        <h1><span class="emoji">🚀</span>Script Server</h1>
        <p>Available script endpoints:</p>
        <div class="click-hint">💡 Click any endpoint to copy the curl command to clipboard</div>
        
%s
        <!-- Rest of the HTML content -->
    </div>
</body>
</html>`, scriptElements.String())
}
