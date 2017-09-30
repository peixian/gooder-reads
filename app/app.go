package app

import (
	"net/http"

	"context"
	"database/sql"
	"fmt"
	"github.com/peixian/gooder-reads/isbn"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type App struct {
	http.ServeMux
	db *sqlx.DB
}

func New(dsn string) (*App, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	m := http.NewServeMux()
	a := App{
		ServeMux: *m,
		db:       db,
	}
	return &a, nil
}

func userFromContext(ctx context.Context) (string, error) {
	user_id, ok := ctx.Value("user").(int)
	if !ok {
		return "", fmt.Errorf("Invalid user context")
	}

	return user_id, nil
}

func shelfFromContext(ctx context.Context) (string, error) {
	shelf, ok := ctx.Value("shelf").(string)
	if !ok {
		return "", fmt.Errof("Invalid shelf context")
	}

	return shelf, nil
}

func (app *App) createBook(w http.ResponseWriter, req *http.Request) {
	isbn := req.URL.Query().Get("ISBN")
	dbBook, err := app.getBookForISBN(isbn)
	if err == sql.ErrNoRows {
		apiBook, err := app.API.Book(req.Context(), isbn)
		if err != nil {
			return err
		}

		book := Book{
			BookName: apiBook.TitleLatin,
			Author:   apiBook.BookAuthor.Name,
			Genre:    apiBook.SubjectID,
			ISBN:     apiBook.ISBN13,
		}
		err = app.insertNewBook(book)
		if err != nil {
			return err
		}
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	user_id, err := userFromContext(req.Context())
	if err != nil {
		return err
	}
	shelf, err := shelfFromContext(req.Context())
	if err != nil {
		return err
	}

	user := User{UserID: user_id}
	shelf := ShelvedBooks{
		UserID:    user_id,
		ISBN:      isbn,
		ShelfName: shelf,
		Progress:  0,
	}

	app.insertNewBookForUser(user, shelf)
}

func (app *App) createBookManually(w http.ResponseWriter, req *http.Request) {
	isbn := req.URL.Query().Get("ISBN")
	bookName := req.URL.Query().Get("BookName")
	author := req.URL.Query().Get("Author")
	genre := req.URL.Query().Get("Genre")

	book := Book{
		BookName: bookName,
		Author:   author,
		ISBN:     isbn,
		Genre:    genre,
	}

	err = app.insertNewBook(book)
	if err != nil {
		return err
	}
	return nil
}
