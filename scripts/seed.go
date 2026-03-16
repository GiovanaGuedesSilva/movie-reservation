//go:build ignore

// Run with: go run ./scripts/seed.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ── Inline GORM models (mirrors secondary/postgres/models.go) ──────────────

type User struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string `gorm:"default:user"`
}

func (User) TableName() string { return "users" }

type Genre struct {
	ID   string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name string `gorm:"uniqueIndex"`
}

func (Genre) TableName() string { return "genres" }

type Movie struct {
	ID              string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Title           string
	Description     string
	PosterURL       string
	DurationMinutes int
	Genres          []Genre `gorm:"many2many:movie_genres;"`
}

func (Movie) TableName() string { return "movies" }

type Theater struct {
	ID          string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string `gorm:"uniqueIndex"`
	TotalRows   int
	SeatsPerRow int
}

func (Theater) TableName() string { return "theaters" }

type Seat struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TheaterID string
	Row       string
	Number    int
}

func (Seat) TableName() string { return "seats" }

// ── Seed ────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "movie_reservation"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	seedAdmin(db)
	genres := seedGenres(db)
	seedMovies(db, genres)
	seedTheaters(db)

	log.Println("seed completed successfully")
}

func seedAdmin(db *gorm.DB) {
	email := getEnv("ADMIN_EMAIL", "admin@cinema.com")
	password := getEnv("ADMIN_PASSWORD", "Admin@123")

	var count int64
	db.Model(&User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		log.Printf("admin %q already exists, skipping", email)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hashing admin password: %v", err)
	}

	admin := User{
		Name:         "Admin",
		Email:        email,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("creating admin: %v", err)
	}

	log.Printf("admin created: %s", email)
}

func seedGenres(db *gorm.DB) []Genre {
	names := []string{"Ação", "Aventura", "Comédia", "Drama", "Ficção Científica"}
	genres := make([]Genre, 0, len(names))

	for _, name := range names {
		g := Genre{Name: name}
		result := db.Where("name = ?", name).FirstOrCreate(&g)
		if result.Error != nil {
			log.Fatalf("seeding genre %q: %v", name, result.Error)
		}
		genres = append(genres, g)
		log.Printf("genre: %s", name)
	}

	return genres
}

func seedMovies(db *gorm.DB, genres []Genre) {
	movies := []Movie{
		{
			Title:           "Duna: Parte 2",
			Description:     "Paul Atreides se une aos Fremen enquanto busca vingança.",
			PosterURL:       "https://image.tmdb.org/t/p/w500/1pdfLvkbY9ohJlCjQH2CZjjYVvJ.jpg",
			DurationMinutes: 166,
			Genres:          []Genre{genres[0], genres[1], genres[4]},
		},
		{
			Title:           "Oppenheimer",
			Description:     "A história do físico J. Robert Oppenheimer e da bomba atômica.",
			PosterURL:       "https://image.tmdb.org/t/p/w500/8Gxv8gSFCU0XGDykEGv7zR1n2ua.jpg",
			DurationMinutes: 180,
			Genres:          []Genre{genres[2], genres[3]},
		},
		{
			Title:           "Barbie",
			Description:     "Barbie e Ken partem para o mundo real após uma crise existencial.",
			PosterURL:       "https://image.tmdb.org/t/p/w500/iuFNMS8vlZuklft5bUMBS2FHPOR.jpg",
			DurationMinutes: 114,
			Genres:          []Genre{genres[2]},
		},
	}

	for _, m := range movies {
		var count int64
		db.Model(&Movie{}).Where("title = ?", m.Title).Count(&count)
		if count > 0 {
			log.Printf("movie %q already exists, skipping", m.Title)
			continue
		}
		if err := db.Create(&m).Error; err != nil {
			log.Fatalf("seeding movie %q: %v", m.Title, err)
		}
		log.Printf("movie created: %s (%d min)", m.Title, m.DurationMinutes)
	}
}

func seedTheaters(db *gorm.DB) {
	theaters := []struct {
		Theater
		Rows int
		Cols int
	}{
		{Theater: Theater{Name: "Sala 1 - IMAX"}, Rows: 10, Cols: 15},
		{Theater: Theater{Name: "Sala 2 - Standard"}, Rows: 8, Cols: 12},
		{Theater: Theater{Name: "Sala 3 - VIP"}, Rows: 5, Cols: 8},
	}

	for _, t := range theaters {
		var count int64
		db.Model(&Theater{}).Where("name = ?", t.Name).Count(&count)
		if count > 0 {
			log.Printf("theater %q already exists, skipping", t.Name)
			continue
		}

		theater := Theater{
			Name:        t.Name,
			TotalRows:   t.Rows,
			SeatsPerRow: t.Cols,
		}

		if err := db.Create(&theater).Error; err != nil {
			log.Fatalf("seeding theater %q: %v", t.Name, err)
		}

		// Generate seats: rows A–Z, columns 1–N
		for row := 0; row < t.Rows; row++ {
			rowLabel := string(rune('A' + row))
			for col := 1; col <= t.Cols; col++ {
				seat := Seat{
					TheaterID: theater.ID,
					Row:       rowLabel,
					Number:    col,
				}
				if err := db.Create(&seat).Error; err != nil {
					log.Fatalf("seeding seat %s%d in theater %q: %v", rowLabel, col, t.Name, err)
				}
			}
		}

		log.Printf("theater created: %s (%d seats)", t.Name, t.Rows*t.Cols)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
