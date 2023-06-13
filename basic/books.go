package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Book struct {
	ID          int
	Title       string
	Author      string
	Publication string
	Borrowed    bool
	Borrower    string
	BorrowedAt  time.Time
}

type Library struct {
	db *sql.DB
}

// function method to add a book
func (l *Library) AddBook(book Book) error {
	var count int
	err := l.db.QueryRow("SELECT COUNT(*) FROM books WHERE id = $1", book.ID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("book with ID %d already exists", book.ID)
	}

	_, err = l.db.Exec("INSERT INTO books (id, title, author, publication, borrowed, borrower, borrowed_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		book.ID, book.Title, book.Author, book.Publication, book.Borrowed, book.Borrower, book.BorrowedAt)
	if err != nil {
		return err
	}
	return nil
}

// function method to remove book
func (l *Library) RemoveBook(bookID int) error {
	result, err := l.db.Exec("DELETE FROM books WHERE id = $1", bookID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("book with ID %d not found", bookID)
	}
	return nil
}

// function method to search books
func (l *Library) SearchBook(keyword string) ([]Book, error) {
	rows, err := l.db.Query("SELECT id, title, author, publication, borrowed, borrower, borrowed_at FROM books WHERE title ILIKE $1 OR author ILIKE $1", "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Publication, &book.Borrowed, &book.Borrower, &book.BorrowedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, book)
	}
	return results, nil
}

// function method to borrow a book
func (l *Library) BorrowBook(bookID int, borrower string) error {
	_, err := l.db.Exec("UPDATE books SET borrowed = true, borrower = $1, borrowed_at = $2 WHERE id = $3", borrower, time.Now(), bookID)
	if err != nil {
		return err
	}
	return nil
}

// function method to return a book
func (l *Library) ReturnBook(bookID int) error {
	_, err := l.db.Exec("UPDATE books SET borrowed = false, borrower = '', borrowed_at = NULL WHERE id = $1", bookID)
	if err != nil {
		return err
	}
	return nil
}

// function to display books
func (l *Library) DisplayBooks() {
	// Fetch all books from the database
	rows, err := l.db.Query("SELECT id, title, author, publication, borrowed, borrower, borrowed_at FROM books")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("All books:")
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Publication, &book.Borrowed, &book.Borrower, &book.BorrowedAt)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("ID: %d, Title: %s, Author: %s, Publication: %s, Borrowed: %t, Borrower: %s, Borrowed At: %s\n",
			book.ID, book.Title, book.Author, book.Publication, book.Borrowed, book.Borrower, book.BorrowedAt)
	}
}

// Command line interface for users to interact with the library
func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=1998 dbname=Library sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	library := Library{db: db}

	library.AddBook(Book{ID: 1, Title: "Book 1", Author: "Author 1", Publication: "Publication 1"})
	library.AddBook(Book{ID: 2, Title: "Book 2", Author: "Author 2", Publication: "Publication 2"})
	library.AddBook(Book{ID: 3, Title: "Book 3", Author: "Author 3", Publication: "Publication 3"})

	for {
		fmt.Println("\n===== Library Management System =====")
		fmt.Println("1. Add a book")
		fmt.Println("2. Remove a book")
		fmt.Println("3. Search for a book")
		fmt.Println("4. Borrow a book")
		fmt.Println("5. Return a book")
		fmt.Println("6. Display all books")
		fmt.Println("0. Exit")
		fmt.Print("Enter your choice: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			var book Book
			fmt.Print("Enter ID: ")
			_, err := fmt.Scanln(&book.ID)
			if err != nil || book.ID <= 0 {
				fmt.Println("Invalid ID. Please enter a positive integer.")
				continue
			}

			reader := bufio.NewReader(os.Stdin)

			for {
				fmt.Print("Enter book title: ")
				title, _ := reader.ReadString('\n')
				title = strings.TrimSpace(title)
				if title == "" {
					fmt.Println("Invalid title. Please enter a non-empty string.")
					continue
				}
				book.Title = title
				break
			}

			for {
				fmt.Print("Enter author name: ")
				author, _ := reader.ReadString('\n')
				author = strings.TrimSpace(author)
				if author == "" {
					fmt.Println("Invalid author. Please enter a non-empty string.")
					continue
				}
				book.Author = author
				break
			}

			for {
				fmt.Print("Enter publication: ")
				publication, _ := reader.ReadString('\n')
				publication = strings.TrimSpace(publication)
				if publication == "" {
					fmt.Println("Invalid publication. Please enter a non-empty string.")
					continue
				}
				book.Publication = publication
				break
			}

			err = library.AddBook(book)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Book added successfully")
			}

		case 2:
			var bookID int
			fmt.Print("Enter book ID: ")
			fmt.Scanln(&bookID)

			err := library.RemoveBook(bookID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book removed successfully")
			}

		case 3:
			var keyword string
			fmt.Print("Enter the keyword to search: ")
			fmt.Scanln(&keyword)

			results, err := library.SearchBook(keyword)
			if err != nil {
				fmt.Println("Error:", err)
			}
			if len(results) == 0 {
				fmt.Println("No matching books found")
			} else {
				fmt.Println("Matching Books:")
				for _, book := range results {
					fmt.Printf("ID: %d, Title: %s, Author: %s, Publication: %s\n", book.ID, book.Title, book.Author, book.Publication)
				}
			}

		case 4:
			var bookID int
			var borrower string
			fmt.Print("Enter book ID: ")
			fmt.Scanln(&bookID)
			fmt.Print("Enter borrower name: ")
			fmt.Scanln(&borrower)

			err := library.BorrowBook(bookID, borrower)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book borrowed successfully")
			}

		case 5:
			var bookID int
			fmt.Print("Enter book ID: ")
			fmt.Scanln(&bookID)

			err := library.ReturnBook(bookID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book returned successfully")
			}

		case 6:
			fmt.Println("All books:")
			library.DisplayBooks()

		case 0:
			fmt.Println("Exiting")
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}
