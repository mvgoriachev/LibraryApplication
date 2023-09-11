package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	model "github.com/iamrajiv/helloworld-grpc-gateway/proto/model"
	"github.com/jackc/pgx/v5"
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
	return
}

func (pgConn *PgConn) ListBooks() (books []*model.Book) {

	rows, err := pgConn.conn.Query(context.Background(), "select * from book")
	if err != nil {
		log.Fatal("error while executing query")
	}

	// iterate through the rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal("error while iterating dataset")
		}
		var book model.Book
		book.Id = values[0].(string)
		book.Title = values[1].(string)
		book.Content = values[2].(string)
		book.CreatedAt = timestamppb.New(values[3].(time.Time))
		book.UpdatedAt = timestamppb.New(values[4].(time.Time))
		books = append(books, &book)
	}
	return

}

func (pgConn *PgConn) HelloQuery() string {
	var greeting string
	err := pgConn.conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return greeting
}
