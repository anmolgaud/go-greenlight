package data

import (
	"time"
	"fmt"
	"greenlight.anmol.gaud/internal/validator"
)

type Movie struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
	Year int32 `json:"year,omitempty"`
	Runtime Runtime `json:"runtime,omitempty"`
	Genres []string `json:"genre,omitempty"`
	Version int32 `json:"version"`
	CreatedAt time.Time `json:"-"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888 && movie.Year <= int32(time.Now().Year()), "year", fmt.Sprintf("year must be between 1888 and %d", time.Now().Year()))

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be positibe integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 value")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 values")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

