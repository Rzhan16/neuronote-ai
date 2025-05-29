package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
)

// Database connection string
const (
	dbConnStr = "postgres://postgres:postgres@postgres:5432/neuronote?sslmode=disable"
	mlBaseURL = "http://ml:8000"
)

// Response models
type Note struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Summary   string     `json:"summary"`
	QuizCards []QuizCard `json:"quiz_cards"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type QuizCard struct {
	ID       string `json:"id"`
	NoteID   string `json:"note_id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type StudyBlock struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	NoteID    string    `json:"note_id"`
	Status    string    `json:"status"`
}

var (
	db       *sql.DB
	mlClient *MLClient
)

// Handlers
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}

func uploadNote(c *fiber.Ctx) error {
	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	// Get user ID from context
	userID := c.Get("X-User-ID")

	// Send to ML service
	noteID, err := mlClient.Pipeline(src, file.Filename, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ML service error: %v", err),
		})
	}

	return c.JSON(fiber.Map{"note_id": noteID})
}

func getNote(c *fiber.Ctx) error {
	noteID := c.Params("id")

	// Get note with quiz cards
	var note Note
	err := db.QueryRow(`
		SELECT id, content, summary, created_at, updated_at
		FROM notes
		WHERE id = $1
	`, noteID).Scan(
		&note.ID,
		&note.Content,
		&note.Summary,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Note not found",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch note",
		})
	}

	// Get quiz cards
	rows, err := db.Query(`
		SELECT id, question, answer
		FROM quiz_cards
		WHERE note_id = $1
	`, noteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quiz cards",
		})
	}
	defer rows.Close()

	for rows.Next() {
		var card QuizCard
		err := rows.Scan(&card.ID, &card.Question, &card.Answer)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan quiz card",
			})
		}
		card.NoteID = noteID
		note.QuizCards = append(note.QuizCards, card)
	}

	return c.JSON(note)
}

func getSchedule(c *fiber.Ctx) error {
	// Get upcoming study blocks
	rows, err := db.Query(`
		SELECT id, note_id, start_time, end_time, status
		FROM study_blocks
		WHERE start_time >= NOW()
		ORDER BY start_time ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch schedule",
		})
	}
	defer rows.Close()

	var blocks []StudyBlock
	for rows.Next() {
		var block StudyBlock
		err := rows.Scan(
			&block.ID,
			&block.NoteID,
			&block.StartTime,
			&block.EndTime,
			&block.Status,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan study block",
			})
		}
		blocks = append(blocks, block)
	}

	return c.JSON(blocks)
}

func main() {
	// Initialize ML client
	mlClient = NewMLClient()

	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Public routes
	app.Get("/health", healthCheck)
	app.Post("/auth/signup", signup)
	app.Post("/auth/login", login)

	// Protected routes
	api := app.Group("/api", authMiddleware())
	api.Post("/notes", uploadNote)
	api.Get("/notes/:id", getNote)
	api.Post("/schedule", createSchedule)
	api.Get("/schedule", getStudySchedule)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
