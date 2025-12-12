package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"greenlight.ilx.net/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func (m *MovieModel) Insert(movie *Movie) error {

	stmt := `INSERT INTO movies (title, year, runtime, genres) VALUES($1, $2, $3, $4) RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)

}

func (m *MovieModel) GetMovie(id int64) (*Movie, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, created_at, title, year, runtime, genres, version  FROM movies WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var movie Movie

	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil

}

func (m *MovieModel) Update(movie *Movie) error {

	stmt := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 WHERE id = $5 AND version = $6 RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
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

func (m *MovieModel) Delete(id int64) error {

	stmt := `DELETE FROM movies WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (m *MovieModel) GetAll(title string, genres []string, f Filters) ([]*Movie, Metadata, error) {
	stmt := fmt.Sprintf(`SELECT count(*) OVER(), * FROM movies WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (genres @> $2 OR $2 = '{}') ORDER BY %s %s, id ASC LIMIT %d OFFSET %d`, f.sortCoulmn(), f.sortDirecetion(), f.Limit(), f.Offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, title, pq.Array(genres))
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var movies []*Movie
	var totalRecords int

	for rows.Next() {
		var movie Movie

		err := rows.Scan(&totalRecords, &movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)

		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)

	}

	metadata := CalculateMetadata(totalRecords, f.Page, f.PageSize)

	return movies, metadata, nil

}

func ValidateMovie(v *validator.Validator, input *Movie) bool {
	v.Check(len(input.Title) > 3, "title", "tetle must be greater than 3 characters")
	v.Check(len(input.Title) < 50, "title", "title must be lower than 50 characters")
	v.Check(validator.Unique(input.Genres), "genres", "you should avoid using duplicate values for genres")
	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	return v.Valid()

}
