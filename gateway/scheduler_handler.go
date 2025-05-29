package main

import (
	"time"

	"neuronote/gateway/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateScheduleRequest struct {
	Notes []struct {
		ID      string    `json:"id"`
		DueDate time.Time `json:"due_date"`
		Weight  float64   `json:"weight"`
	} `json:"notes"`
	Calendar []struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
		Busy  bool      `json:"busy"`
	} `json:"calendar"`
}

func createSchedule(c *fiber.Ctx) error {
	// Parse request
	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if len(req.Notes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one note is required",
		})
	}
	if len(req.Calendar) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Calendar slots are required",
		})
	}

	// Convert request to solver input
	notes := make([]scheduler.Note, len(req.Notes))
	for i, n := range req.Notes {
		notes[i] = scheduler.Note{
			ID:      n.ID,
			DueDate: n.DueDate,
			Weight:  n.Weight,
		}
	}

	calendar := make([]scheduler.CalendarSlot, len(req.Calendar))
	for i, s := range req.Calendar {
		calendar[i] = scheduler.CalendarSlot{
			Start: s.Start,
			End:   s.End,
			Busy:  s.Busy,
		}
	}

	// Get user ID from context
	userID := c.Get("X-User-ID")

	// Create solver and generate schedule
	solver := scheduler.NewSolver(notes, calendar, userID)
	blocks, err := solver.Solve()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate schedule",
		})
	}

	// Save blocks to database
	tx, err := db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start transaction",
		})
	}
	defer tx.Rollback()

	for i := range blocks {
		blocks[i].ID = uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO study_blocks (id, user_id, note_id, start_time, end_time)
			VALUES ($1, $2, $3, $4, $5)
		`, blocks[i].ID, blocks[i].UserID, blocks[i].NoteID, blocks[i].Start, blocks[i].End)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save study blocks",
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(blocks)
}

func getStudySchedule(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Get("X-User-ID")

	// Get upcoming study blocks
	rows, err := db.Query(`
		SELECT id, note_id, start_time, end_time
		FROM study_blocks
		WHERE user_id = $1 AND start_time >= NOW()
		ORDER BY start_time ASC
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch schedule",
		})
	}
	defer rows.Close()

	var blocks []scheduler.StudyBlock
	for rows.Next() {
		var block scheduler.StudyBlock
		err := rows.Scan(
			&block.ID,
			&block.NoteID,
			&block.Start,
			&block.End,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan study block",
			})
		}
		block.UserID = userID
		blocks = append(blocks, block)
	}

	return c.JSON(blocks)
}
