package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	model "github.com/mvgoriachev/LibraryApplication/proto/model"
	services "github.com/mvgoriachev/LibraryApplication/proto/services"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

//var conn *pgx.Conn

type PgConn struct {
	conn *pgx.Conn
}

const dbConnString = "postgresql://postgres:netcracker@localhost:5432/MyLibrary"

func GetConnection() *PgConn {
	conn, err := pgx.Connect(context.Background(), dbConnString)
	if err != nil {
		log.Fatal("Unable to connect to database: %v\n", err)
	}
	//defer conn.Close(context.Background())
	pgConn := PgConn{conn}
	return &pgConn
}

func (pgConn *PgConn) GetBook(id string) (book model.Book) {
	var t1, t2 time.Time
	err := pgConn.conn.QueryRow(context.Background(), "select * from book where book_id = ($1)", id).Scan(&book.Id, &book.Title, &book.Content, &t1, &t2)
	if err != nil {
		log.Fatal("QueryRow failed: %v\n", err)
	}
	book.CreatedAt = timestamppb.New(t1)
	book.UpdatedAt = timestamppb.New(t2)

	rows, err2 := pgConn.conn.Query(context.Background(), "select * from author where book_id = ($1)", id)
	if err2 != nil {
		log.Fatal("QueryRow failed: %v\n", err2)
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal("error while iterating dataset")
		}
		book.Authors = append(book.Authors, values[1].(string))
	}
	return
}

func (pgConn *PgConn) ListBooks() (books []*model.Book) {

	rows, err := pgConn.conn.Query(context.Background(), "select * from book")
	if err != nil {
		log.Fatal("error while executing query")
	}

	var bookIds []string
	// iterate through the rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal("error while iterating dataset")
		}

		//x, y := values[0].(string)
		//log.Println(x, " ", y)
		var book model.Book
		book.Id = fmt.Sprint(values[0].(int32))
		book.Title = values[1].(string)
		book.Content = values[2].(string)
		book.CreatedAt = timestamppb.New(values[3].(time.Time))
		book.UpdatedAt = timestamppb.New(values[4].(time.Time))
		books = append(books, &book)
		bookIds = append(bookIds, book.Id)
	}

	//collect authors for needed book IDs only
	rows, err = pgConn.conn.Query(context.Background(), "select * from author where book_id = any ($1)", bookIds)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal("error while iterating dataset")
		}
		for _, book := range books {
			if book.Id == fmt.Sprint(values[0].(int32)) {
				book.Authors = append(book.Authors, values[1].(string))
			}
		}
	}

	return

}

func (pgConn *PgConn) CreateBook(bookReq services.CreateRequest) (bookResp model.Book) {
	var t1, t2 time.Time
	var id int
	updateBookQuery := `INSERT INTO book (book_title, book_content, created_at, updated_at)
	VALUES ($1, $2, current_timestamp, current_timestamp)
	RETURNING book_id, created_at, updated_at;`
	err := pgConn.conn.QueryRow(context.Background(), updateBookQuery, bookReq.Title, bookReq.Content).Scan(&id, &t1, &t2)
	if err != nil {
		log.Fatal("QueryRow failed: %v\n", err)
	}
	bookResp.Id = id
	bookResp.CreatedAt = timestamppb.New(t1)
	bookResp.UpdatedAt = timestamppb.New(t2)

	rows := [][]interface{}{
		{"1", "Smith"},
		{"1", "Doe"},
	}

	_, err2 := pgConn.conn.CopyFrom(
		pgx.Identifier{"author"},
		[]string{"book_id", "author_name"},
		pgx.CopyFromRows(rows),
	)
	if err2 != nil {
		log.Fatal("QueryRow failed: %v\n", err2)
	}

	return
}
