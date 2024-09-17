package main

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host      = "localhost"
	port      = 5432
	user      = "username"
	password  = "password"
	dbname    = "default_database"
	tableName = "seats"
	seatCount = 1000
)

func connect(ctx context.Context) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(2)

	// Drop the table if it exists
	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS flight_bookings`)
	if err != nil {
		log.Fatal(err)
	}

	// Create flight booking table
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS flight_bookings (
	id SERIAL PRIMARY KEY,
	booked_by VARCHAR(255) UNIQUE
)`)

	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.PrepareContext(ctx, "INSERT INTO flight_bookings (booked_by) VALUES ($1)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for range seatCount {
		_, err = stmt.ExecContext(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Added entries to flight_bookings table.")
	return db
}

type Seat struct {
}

func show(ctx context.Context) {
}

func main() {
	ctx := context.Background()
	db := connect(ctx)
	defer db.Close()

	wg := &sync.WaitGroup{}
	id := atomic.Int32{}
	book := func(count int) {
		defer wg.Done()
		conn, err := db.Conn(ctx)
		if err != nil {
			log.Println("Error getting connection: %v", err)
			return
		}
		defer conn.Close()

		innerWg := &sync.WaitGroup{}
		defer innerWg.Wait()

		for range count {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				i := id.Add(1)
				_, err = conn.ExecContext(ctx, `UPDATE flight_bookings
	SET booked_by = $1
	WHERE id = (
		SELECT id
		FROM flight_bookings
		WHERE booked_by IS NULL
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	)`, i)
				if err != nil {
					log.Printf("Error booking seat: %v", err)
					return
				}
			}()
		}
	}

	countBookings := func() {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM flight_bookings WHERE booked_by IS NOT NULL").Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Total bookings: %d\n", count)
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM flight_bookings WHERE booked_by IS NULL").Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Total seats left: %d\n", count)
	}

	start := time.Now()

	threads := 50
	for range threads {
		wg.Add(1)
		go book(seatCount / threads)
	}

	fmt.Println("waiting")
	wg.Wait()
	fmt.Println("done")
	fmt.Println(time.Since(start))

	countBookings()
}
