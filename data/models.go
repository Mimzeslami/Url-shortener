package data

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		Url: Url{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	Url Url
}

// User is the structure which holds one user from the database.
type Url struct {
	gorm.Model
	ShortUrl string `gorm:"unique;not null"`
	LongUrl  string
}
