package main

import (
	"errors"
	"net/http"
	"oynas/internal/data"
	"oynas/internal/validator"
)

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text   string      `json:"text"`
		Rating data.Rating `json:"rating"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	comment := &data.Comment{
		ToyID:    id,
		UserName: user.Name,
		Text:     input.Text,
		Rating:   input.Rating,
	}

	v := validator.New()

	if data.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Comment.Insert(comment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateComment):
			v.AddError("email", "comment already exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
