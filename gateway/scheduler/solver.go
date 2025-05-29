package scheduler

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Note represents a study note with scheduling requirements
type Note struct {
	ID      string    `json:"id"`
	DueDate time.Time `json:"due_date"`
	Weight  float64   `json:"weight"`
}

// CalendarSlot represents a time slot in the user's calendar
type CalendarSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Busy  bool      `json:"busy"`
}

// StudyBlock represents a scheduled study session
type StudyBlock struct {
	ID     string    `json:"id"`
	UserID string    `json:"user_id"`
	NoteID string    `json:"note_id"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

// Constants for retention calculation (Ebbinghaus forgetting curve)
const (
	retentionStrength = 1.84 // Strength of memory
	minimumRetention  = 0.4  // Minimum acceptable retention
	studyBoost        = 0.3  // Retention boost from study session
	slotDuration      = 30   // Duration of each study slot in minutes
	maxSlotsPerNote   = 3    // Maximum study sessions per note
)

// calculateRetention returns the retention probability at a given time before due date
// Based on Ebbinghaus forgetting curve: R = e^(-t/S) where:
// R = retention probability
// t = time since last review (in days)
// S = strength of memory
func calculateRetention(timeBeforeDue time.Duration) float64 {
	days := timeBeforeDue.Hours() / 24
	return math.Exp(-days / retentionStrength)
}

// Solver creates an optimal study schedule
type Solver struct {
	notes    []Note
	calendar []CalendarSlot
	userID   string
}

// NewSolver creates a new scheduler solver
func NewSolver(notes []Note, calendar []CalendarSlot, userID string) *Solver {
	return &Solver{
		notes:    notes,
		calendar: calendar,
		userID:   userID,
	}
}

// Solve generates an optimal study schedule using a greedy algorithm
func (s *Solver) Solve() ([]StudyBlock, error) {
	// Convert calendar slots to discrete 30-minute intervals
	slots := s.discretizeCalendar()
	if len(slots) == 0 {
		return nil, fmt.Errorf("no available time slots")
	}

	// Sort notes by priority (weight * urgency)
	type noteWithPriority struct {
		note     Note
		priority float64
	}
	notePriorities := make([]noteWithPriority, len(s.notes))
	for i, note := range s.notes {
		timeUntilDue := note.DueDate.Sub(time.Now())
		urgency := 1.0 / (timeUntilDue.Hours()/24 + 1) // +1 to avoid division by zero
		notePriorities[i] = noteWithPriority{
			note:     note,
			priority: note.Weight * urgency,
		}
	}
	sort.Slice(notePriorities, func(i, j int) bool {
		return notePriorities[i].priority > notePriorities[j].priority
	})

	// Allocate slots to notes
	var blocks []StudyBlock
	noteCount := make(map[string]int)

	for _, np := range notePriorities {
		note := np.note
		if noteCount[note.ID] >= maxSlotsPerNote {
			continue
		}

		// Find best available slots for this note
		bestSlots := s.findBestSlots(note, slots, noteCount[note.ID])
		if len(bestSlots) == 0 {
			continue
		}

		// Create study blocks
		for _, slot := range bestSlots {
			blocks = append(blocks, StudyBlock{
				UserID: s.userID,
				NoteID: note.ID,
				Start:  slot.Start,
				End:    slot.End,
			})
			noteCount[note.ID]++

			// Mark slot as busy
			for i := range slots {
				if slots[i].Start == slot.Start {
					slots[i].Busy = true
				}
			}
		}
	}

	return blocks, nil
}

// findBestSlots finds the best available slots for a note based on retention
func (s *Solver) findBestSlots(note Note, slots []CalendarSlot, existingSlots int) []CalendarSlot {
	slotsNeeded := maxSlotsPerNote - existingSlots
	if slotsNeeded <= 0 {
		return nil
	}

	// Score each available slot based on retention
	type scoredSlot struct {
		slot  CalendarSlot
		score float64
	}
	var scoredSlots []scoredSlot

	for _, slot := range slots {
		if slot.Busy {
			continue
		}

		timeBeforeDue := note.DueDate.Sub(slot.End)
		if timeBeforeDue <= 0 {
			continue
		}

		retention := calculateRetention(timeBeforeDue)
		score := retention * note.Weight
		scoredSlots = append(scoredSlots, scoredSlot{
			slot:  slot,
			score: score,
		})
	}

	// Sort slots by score
	sort.Slice(scoredSlots, func(i, j int) bool {
		return scoredSlots[i].score > scoredSlots[j].score
	})

	// Take the best slots
	var bestSlots []CalendarSlot
	for i := 0; i < len(scoredSlots) && i < slotsNeeded; i++ {
		bestSlots = append(bestSlots, scoredSlots[i].slot)
	}

	return bestSlots
}

// discretizeCalendar converts calendar slots into 30-minute intervals
func (s *Solver) discretizeCalendar() []CalendarSlot {
	var slots []CalendarSlot
	for _, slot := range s.calendar {
		current := slot.Start
		for current.Before(slot.End) {
			end := current.Add(time.Duration(slotDuration) * time.Minute)
			if end.After(slot.End) {
				end = slot.End
			}
			slots = append(slots, CalendarSlot{
				Start: current,
				End:   end,
				Busy:  slot.Busy,
			})
			current = end
		}
	}
	return slots
}
