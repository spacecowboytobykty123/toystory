package main

import (
	"errors"
	"fmt"
	"net/http"
	"oynas/internal/data"
	"oynas/internal/validator"
)

func (app *application) createToyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title          string   `json:"title"`
		Description    string   `json:"desc"`
		Details        []string `json:"details"`
		Skills         []string `json:"skills"`
		Images         []string `json:"images"`
		Categories     []string `json:"categories"`
		RecommendedAge string   `json:"recAge"`
		Manufacturer   string   `json:"manufac"`
		Value          int64    `json:"value"`
		IsAvailable    string   `json:"isAvailable"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	toy := &data.Toy{
		Title:          input.Title,
		Description:    input.Description,
		Details:        input.Details,
		Skills:         input.Skills,
		Images:         input.Images,
		Categories:     input.Categories,
		RecommendedAge: input.RecommendedAge,
		Manufacturer:   input.Manufacturer,
		Value:          input.Value,
		IsAvailable:    input.IsAvailable,
	}

	v := validator.New()

	if data.ValidateToy(v, toy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Toys.Insert(toy)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/toys/%d", toy.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"toy": toy}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showToyHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Text   string      `json:"text"`
		Rating data.Rating `json:"rating"`
	}

	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	toy, err := app.models.Toys.Get(id)

	comments, err := app.models.Comment.GetCommentsFromId(id, input.Text, input.Rating)

	err = app.writeJSON(w, http.StatusOK, envelope{"toy": toy, "comments": comments}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateToyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	toy, err := app.models.Toys.Get(id)
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
		Title          *string   `json:"title"`
		Description    *string   `json:"desc"`
		Details        *[]string `json:"details"`
		Skills         *[]string `json:"skills"`
		Categories     *[]string `json:"categories"`
		RecommendedAge *string   `json:"recAge"`
		Manufacturer   *string   `json:"manufacturer"`
		Value          *int64    `json:"value"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		toy.Title = *input.Title
	}
	if input.Description != nil {
		toy.Description = *input.Description
	}
	if input.Details != nil {
		toy.Details = *input.Details
	}
	if input.Skills != nil {
		toy.Skills = *input.Skills
	}
	if input.Categories != nil {
		toy.Categories = *input.Categories
	}
	if input.RecommendedAge != nil {
		toy.RecommendedAge = *input.RecommendedAge
	}
	if input.Manufacturer != nil {
		toy.Manufacturer = *input.Manufacturer
	}
	if input.Value != nil {
		toy.Value = *input.Value
	}

	v := validator.New()
	if data.ValidateToy(v, toy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Toys.Update(toy)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"toy": toy}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteToyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Toys.Delete(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Toy deleted successfully"}, nil)

}

func (app *application) listToysHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title          string
		Value          int64
		From           int64
		To             int64
		Skills         []string
		Categories     []string
		RecommendedAge string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Skills = app.readCSV(qs, "skills", []string{})
	input.Categories = app.readCSV(qs, "categories", []string{})
	input.Value = int64(app.readInt(qs, "value", 0, v))
	input.From = int64(app.readInt(qs, "from", 0, v))
	input.To = int64(app.readInt(qs, "to", 100000, v))

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "skills", "categories", "recAge", "value", "from", "to", "-id", "-title", "-skills", "-categories", "-recAge", "-value"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	toys, metadata, err := app.models.Toys.GetAll(input.Title, input.Skills, input.Categories, input.Value, input.From, input.To, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"toys": toys, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
