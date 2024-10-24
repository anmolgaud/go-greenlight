package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
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

type MovieModel struct {
	DB *sql.DB
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

func (m MovieModel) Insert(movie *Movie) error {
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version;`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	query := `
	SELECT id, title, year, runtime, genres, version, created_at
	FROM movies
	WHERE movies.id = $1;`

	movie := Movie{}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
		&movie.CreatedAt,
	);
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil;
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version;`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	query := `
	DELETE FROM movies
	WHERE id = $1;`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id);
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// used overlap(&&) operator instead of contains(@>) operator
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, title, year, runtime, genres, version, created_at
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres && $2 OR $2 = '{}')
	ORDER By %s %s, id ASC
	LIMIT $3 OFFSET $4;`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	args := []any{
		title,
		pq.Array(genres),
		filters.limit(),
		filters.offset(),
	}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	movies := []*Movie{}
	for rows.Next() {
		movie := Movie{}
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
			&movie.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return movies, metadata, nil
}