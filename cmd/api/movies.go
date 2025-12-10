package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.ilx.net/internal/data"
	"greenlight.ilx.net/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	v := validator.New()

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	if !data.ValidateMovie(v, movie) {
		app.failedValidationResponse(w, r, v.Errors)
		fmt.Fprintln(w, input.Title)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	headers := make(http.Header)

	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, r, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIdParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.GetMovie(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIdParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.GetMovie(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if input.Title != "" {
		movie.Title = input.Title
	}
	if input.Year != 0 {
		movie.Year = input.Year
	}
	if input.Runtime != 0 {
		movie.Runtime = input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIdParam(w, r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"message": "The movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
