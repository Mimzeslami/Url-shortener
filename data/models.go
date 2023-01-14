package data

import (
	"context"
	"database/sql"
	"log"
	"time"
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

type Url struct {
	ID        int       `json:"id"`
	LongUrl   string    `json:"long_url"`
	ShortUrl  string    `json:"short_url" binding:"required,short_url" gorm:"unique,not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *Url) GetAll() ([]*Url, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, long_url,short_url, created_at, updated_at
	from urls order by created_at`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*Url

	for rows.Next() {
		var url Url
		err := rows.Scan(
			&url.ID,
			&url.ShortUrl,
			&url.LongUrl,
			&url.CreatedAt,
			&url.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		urls = append(urls, &url)
	}

	return urls, nil
}

func (u *Url) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from urls where id = $1`

	_, err := db.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (u *Url) Insert(url Url) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `insert into urls (short_url, long_url,created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	err := db.QueryRowContext(ctx, stmt,
		url.ShortUrl,
		url.LongUrl,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (u *Url) GetByShortUrl(shortUrl string) (*Url, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, short_url, long_url, created_at, updated_at from urls where short_url = $1`

	var url Url
	row := db.QueryRowContext(ctx, query, shortUrl)

	err := row.Scan(
		&url.ID,
		&url.ShortUrl,
		&url.LongUrl,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}
