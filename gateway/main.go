package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
)

// Database connection string
const (
	mlBaseURL = "http://ml:8000"
)

// Response models
type Note struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
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
	db *sql.DB
)

// Handlers
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}

func getNotes(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Get("X-User-ID")

	// Query notes from database
	rows, err := db.Query(`
		SELECT n.id, n.title, n.content, n.summary, n.created_at, n.updated_at,
			   array_agg(json_build_object(
				   'id', q.id,
				   'question', q.question,
				   'answer', q.answer
			   )) as quiz_cards
		FROM notes n
		LEFT JOIN quiz_cards q ON n.id = q.note_id
		WHERE n.user_id = $1
		GROUP BY n.id
		ORDER BY n.created_at DESC
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch notes",
		})
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Summary, &note.CreatedAt, &note.UpdatedAt, &note.QuizCards)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan note",
			})
		}
		notes = append(notes, note)
	}

	return c.JSON(notes)
}

func getStudyBlocks(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Get("X-User-ID")

	// Query study blocks from database
	rows, err := db.Query(`
		SELECT id, start_time, end_time, note_id, status
		FROM study_blocks
		WHERE user_id = $1
		ORDER BY start_time ASC
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch study blocks",
		})
	}
	defer rows.Close()

	var blocks []StudyBlock
	for rows.Next() {
		var block StudyBlock
		err := rows.Scan(&block.ID, &block.StartTime, &block.EndTime, &block.NoteID, &block.Status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan study block",
			})
		}
		blocks = append(blocks, block)
	}

	return c.JSON(blocks)
}

func uploadNote(c *fiber.Ctx) error {
	// Parse the uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] No file uploaded: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	log.Printf("[INFO] Received file: %s (%d bytes, header type: %s)", fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"))

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("[ERROR] Failed to open file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer file.Close()

	// Read the file into a buffer (for forwarding or processing)
	var buf bytes.Buffer
	size, err := io.Copy(&buf, file)
	if err != nil || size == 0 {
		log.Printf("[ERROR] File is empty or could not be read: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File is empty or could not be read",
		})
	}
	log.Printf("[INFO] File read into buffer: %d bytes", size)

	// Get user ID from context
	userID := c.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous" // Fallback for testing
	}
	log.Printf("[INFO] User ID: %s", userID)

	// Forward to ML service as multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		log.Printf("[ERROR] Failed to create form file for ML service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create form file for ML service",
		})
	}
	if _, err = io.Copy(part, bytes.NewReader(buf.Bytes())); err != nil {
		log.Printf("[ERROR] Failed to copy file data to ML service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to copy file data to ML service",
		})
	}
	if err = writer.WriteField("user_id", userID); err != nil {
		log.Printf("[ERROR] Failed to add user ID to ML service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add user ID to ML service",
		})
	}
	writer.Close()

	mlReq, err := http.NewRequest("POST", mlBaseURL+"/pipeline", body)
	if err != nil {
		log.Printf("[ERROR] Failed to create ML request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create ML request",
		})
	}
	mlReq.Header.Set("Content-Type", writer.FormDataContentType())

	log.Printf("[INFO] Sending file to ML service at %s/pipeline", mlBaseURL)
	mlResp, err := http.DefaultClient.Do(mlReq)
	if err != nil {
		log.Printf("[ERROR] ML service error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ML service error: " + err.Error(),
		})
	}
	defer mlResp.Body.Close()

	if mlResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(mlResp.Body)
		log.Printf("[ERROR] ML service returned status %d: %s", mlResp.StatusCode, string(body))
		return c.Status(mlResp.StatusCode).JSON(fiber.Map{
			"error": "ML service error: " + string(body),
		})
	}

	// Parse ML service response
	var mlResult struct {
		NoteID string `json:"note_id"`
	}
	if err := json.NewDecoder(mlResp.Body).Decode(&mlResult); err != nil {
		log.Printf("[ERROR] Failed to parse ML response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse ML response: " + err.Error(),
		})
	}
	log.Printf("[INFO] ML service returned note_id: %s", mlResult.NoteID)

	// Return the created note
	var note Note
	err = db.QueryRow(`
		SELECT n.id, n.title, n.content, n.summary, n.created_at, n.updated_at,
			   array_agg(json_build_object(
				   'id', q.id,
				   'question', q.question,
				   'answer', q.answer
			   )) as quiz_cards
		FROM notes n
		LEFT JOIN quiz_cards q ON n.id = q.note_id
		WHERE n.id = $1
		GROUP BY n.id
	`, mlResult.NoteID).Scan(&note.ID, &note.Title, &note.Content, &note.Summary, &note.CreatedAt, &note.UpdatedAt, &note.QuizCards)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch created note: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch created note",
		})
	}
	log.Printf("[INFO] Successfully uploaded and processed note: %s", note.ID)

	return c.JSON(note)
}

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-User-ID",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		ExposeHeaders:    "Content-Length, Content-Type",
		AllowCredentials: true,
	}))

	// Routes
	app.Get("/health", healthCheck)
	app.Get("/notes", getNotes)
	app.Get("/study-blocks", getStudyBlocks)
	app.Post("/notes", uploadNote)
	app.Post("/notes/upload", uploadNote)

	// Start server
	log.Fatal(app.Listen(":8080"))
}
