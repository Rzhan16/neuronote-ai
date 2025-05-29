package main

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSecret    = "dev_secret"
	tokenExpiry  = 30 * 24 * time.Hour // 30 days
	cookieName   = "nn_token"
	cookieMaxAge = 30 * 24 * 60 * 60 // 30 days in seconds
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func signup(c *fiber.Ctx) error {
	var req SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	// Create user
	var userID string
	err = db.QueryRow(
		"INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING id",
		req.Email,
		string(hashedPassword),
	).Scan(&userID)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate JWT token
	token, err := generateToken(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Store session
	expiresAt := time.Now().Add(tokenExpiry)
	_, err = db.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID,
		token,
		expiresAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   cookieMaxAge,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user_id": userID,
		"token":   token,
	})
}

func login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user
	var user struct {
		ID             string
		HashedPassword string
	}
	err := db.QueryRow(
		"SELECT id, hashed_password FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.HashedPassword)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate JWT token
	token, err := generateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Store session
	expiresAt := time.Now().Add(tokenExpiry)
	_, err = db.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)",
		user.ID,
		token,
		expiresAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   cookieMaxAge,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id": user.ID,
		"token":   token,
	})
}

func generateToken(userID string) (string, error) {
	claims := Claims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from cookie
		token := c.Cookies(cookieName)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authentication",
			})
		}

		// Verify token
		claims := &Claims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Check session
		var expiresAt time.Time
		err = db.QueryRow(
			"SELECT expires_at FROM sessions WHERE token = $1 AND user_id = $2",
			token,
			claims.UserID,
		).Scan(&expiresAt)
		if err == sql.ErrNoRows || time.Now().After(expiresAt) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session expired",
			})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify session",
			})
		}

		// Set user ID in context
		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
