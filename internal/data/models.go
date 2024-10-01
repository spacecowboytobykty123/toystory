package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Toys        ToyModel
	Permissions PermissionModel
	Users       UserModel
	Comment     CommentModel
	Tokens      TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Toys:        ToyModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Users:       UserModel{DB: db},
		Comment:     CommentModel{DB: db},
		Tokens:      TokenModel{DB: db},
	}
}
