package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSolver_Solve(t *testing.T) {
	// Test data
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	dayAfter := now.Add(48 * time.Hour)

	notes := []Note{
		{
			ID:      "note1",
			DueDate: tomorrow,
			Weight:  1.0,
		},
		{
			ID:      "note2",
			DueDate: dayAfter,
			Weight:  0.8,
		},
	}

	calendar := []CalendarSlot{
		{
			Start: now,
			End:   now.Add(2 * time.Hour),
			Busy:  false,
		},
		{
			Start: now.Add(2 * time.Hour),
			End:   now.Add(3 * time.Hour),
			Busy:  true, // Busy slot
		},
		{
			Start: now.Add(3 * time.Hour),
			End:   now.Add(5 * time.Hour),
			Busy:  false,
		},
	}

	solver := NewSolver(notes, calendar, "test-user")
	blocks, err := solver.Solve()

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, blocks)

	// Check that blocks are within calendar slots
	for _, block := range blocks {
		assert.True(t, block.Start.After(now) || block.Start.Equal(now))
		assert.True(t, block.End.Before(now.Add(5*time.Hour)) || block.End.Equal(now.Add(5*time.Hour)))
		assert.Equal(t, "test-user", block.UserID)
		assert.Contains(t, []string{"note1", "note2"}, block.NoteID)

		// Check block duration
		duration := block.End.Sub(block.Start)
		assert.Equal(t, time.Duration(slotDuration)*time.Minute, duration)

		// Check that block is not in busy slot
		assert.False(t, (block.Start.Equal(now.Add(2*time.Hour)) || block.Start.After(now.Add(2*time.Hour))) &&
			(block.End.Equal(now.Add(3*time.Hour)) || block.End.Before(now.Add(3*time.Hour))))
	}

	// Check that each note has at most maxSlotsPerNote blocks
	noteCount := make(map[string]int)
	for _, block := range blocks {
		noteCount[block.NoteID]++
		assert.LessOrEqual(t, noteCount[block.NoteID], maxSlotsPerNote)
	}

	// Check that note1 (higher priority) has more slots than note2
	if noteCount["note1"] > 0 && noteCount["note2"] > 0 {
		assert.GreaterOrEqual(t, noteCount["note1"], noteCount["note2"])
	}
}

func TestSolver_NoSolution(t *testing.T) {
	// Test with no available slots
	now := time.Now()
	notes := []Note{
		{
			ID:      "note1",
			DueDate: now.Add(24 * time.Hour),
			Weight:  1.0,
		},
	}

	calendar := []CalendarSlot{
		{
			Start: now,
			End:   now.Add(2 * time.Hour),
			Busy:  true, // All slots are busy
		},
	}

	solver := NewSolver(notes, calendar, "test-user")
	blocks, err := solver.Solve()

	assert.NoError(t, err) // Changed from Error to NoError since empty solution is valid
	assert.Empty(t, blocks)
}

func TestCalculateRetention(t *testing.T) {
	tests := []struct {
		name         string
		timeToReview time.Duration
		expected     float64
	}{
		{
			name:         "immediate review",
			timeToReview: 0,
			expected:     1.0,
		},
		{
			name:         "one day",
			timeToReview: 24 * time.Hour,
			expected:     0.58, // approximate
		},
		{
			name:         "one week",
			timeToReview: 7 * 24 * time.Hour,
			expected:     0.02, // approximate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retention := calculateRetention(tt.timeToReview)
			assert.InDelta(t, tt.expected, retention, 0.1)
		})
	}
}

func TestFindBestSlots(t *testing.T) {
	now := time.Now()
	note := Note{
		ID:      "note1",
		DueDate: now.Add(48 * time.Hour),
		Weight:  1.0,
	}

	slots := []CalendarSlot{
		{
			Start: now,
			End:   now.Add(30 * time.Minute),
			Busy:  false,
		},
		{
			Start: now.Add(30 * time.Minute),
			End:   now.Add(60 * time.Minute),
			Busy:  true,
		},
		{
			Start: now.Add(60 * time.Minute),
			End:   now.Add(90 * time.Minute),
			Busy:  false,
		},
	}

	solver := NewSolver([]Note{note}, slots, "test-user")
	bestSlots := solver.findBestSlots(note, slots, 0)

	assert.NotEmpty(t, bestSlots)
	assert.Equal(t, 2, len(bestSlots)) // Should find 2 available slots
	for _, slot := range bestSlots {
		assert.False(t, slot.Busy)
	}
}
