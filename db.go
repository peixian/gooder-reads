package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE books (
    book_name text,
    isbn isbn13,
    author text,
    genre text,
    pages integer
);

CREATE TABLE shelf_books (
    user_id integer,
    isbn isbn13,
    shelf_name text,
    pages_read integer
);

CREATE TABLE shelves (
    user_id integer,
    shelf_name text
);


CREATE TABLE tags (
    user_id integer,
    isbn isbn13,
    tags text[]
);


CREATE TABLE users (
    user_id SERIAL,
    password bytea,
    user_name text,
)
`

type (
	Book struct {
		BookName string `db:"book_name"`
		ISBN     string `db:"isbn"`
		Author   string `db:"author"`
		Genre    string `db:"genre"`
		Pages    int    `db:"pages"`
	}

	ShelvedBooks struct {
		UserID    int    `db:"user_id"`
		ISBN      string `db:"isbn"`
		ShelfName string `db:"shelf_name"`
		PagesRead int    `db:"pages_read"`
	}

	Shelf struct {
		UserID    int    `db:"user_id"`
		ShelfName string `db:"shelf_name"`
	}

	Tags struct {
		UserID int      `db:"user_id"`
		ISBN   string   `db:"isbn"`
		Tags   []string `db:"tags"`
	}

	User struct {
		UserID   int    `db:"user_id"`
		Password string `db:"password"`
	}

	Bookshelf struct {
		Book
		ShelvedBooks
	}
)

func main() {
	books, err := getBooksForUser(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", books[0].ShelvedBooks.ISBN)
	fmt.Printf("%+v", books)

	fmt.Println(setupUser("test"))

}

var db, err = sqlx.Connect("postgres", "user=peixianwang dbname=postgres sslmode=disable")

func getBooksForUser(user int) ([]Bookshelf, error) {
	bookshelves := []Bookshelf{}
	err = db.Select(&bookshelves, "SELECT books.*, shelf_books.shelf_name, shelf_books.pages_read FROM shelf_books JOIN books on books.isbn = shelf_books.isbn WHERE user_id= $1", user)
	if err != nil {
		fmt.Println(err)
		return bookshelves, fmt.Errorf("Error finding user w/ ID: %v", user)
	}

	return bookshelves, nil
}

func getShelvesForUser(user int) ([]ShelvedBooks, error) {
	shelves := []ShelvedBooks{}
	err = db.Select(&shelves, "SELECT * from shelves WHERE user_id = $1", user)
	if err != nil {
		fmt.Println(err)
		return shelves, fmt.Errorf("Error finding user w/ ID: %v", user)
	}

	return shelves, nil
}

func getBookForISBN(isbn string) (Book, error) {
	book := Book{}
	err = db.Get(&book, "SELECT * FROM books WHERE isbn = $1", isbn)
	if err != nil {
		fmt.Println(err)
		return book, fmt.Errorf("Error finding book with ISBN: %v", isbn)
	}

	return book, nil
}

func setupUser(password string) error {
	user := User{Password: password}
	err := db.Get(&user, "INSERT INTO users (password) VALUES ($1) RETURNING user_id", password)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println()
	fmt.Println(user.UserID)
	setup := `INSERT INTO shelves (shelf_name, user_id) VALUES
('currently-reading', $1),
('finished', $1),
('to-read', $1);
`
	_, err = db.Exec(setup, user.UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
