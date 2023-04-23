package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const rowCount = 50

const (
	host      = "localhost"
	port      = 5432
	user      = "username"
	password  = "password"
	dbname    = "default_database"
	tableName = "seats"
)

func connect(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	cfg, err := pgxpool.ParseConfig(connStr)
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 5 * time.Minute
	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type Seat struct {
	Id     int    `db:"id"`
	Row    int    `db:"row"`
	Col    string `db:"col"`
	UserId int    `db:"user_id"`
}

func show(ctx context.Context, conn *pgxpool.Pool) {
	for _, col := range []string{"a", "b", "c", "_", "d", "e", "f"} {
		fmt.Println()
		if col == "_" {
			continue
		}
		for row := 0; row < rowCount; row++ {
			qr, err := conn.Query(ctx, "SELECT * FROM seats WHERE col=$1 AND row=$2", col, row)
			if err != nil {
				log.Error().Err(err).Msg("failed to fetch row")
				qr.Close()
				continue
			}
			seats, err := pgx.CollectRows(qr, pgx.RowToStructByName[Seat])
			if err != nil {
				log.Error().Err(err).Msg("failed to collect")
			} else {
				// fmt.Printf("%03d-", seats[0].UserId)
				if seats[0].UserId == 0 {
					fmt.Print(".")
				} else {
					fmt.Print("x")
				}
			}
			qr.Close()
		}
	}
	fmt.Println()
}

func main() {
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	conn, err := connect(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect")
	}

	t, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS seats (
		id SERIAL PRIMARY KEY,
		row INT NOT NULL,
		col VARCHAR NOT NULL,
		user_id INT DEFAULT 0,
		UNIQUE (row, col)
	);`)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create table")
	} else {
		log.Info().Msg(t.String())
	}

	for row := 0; row < rowCount; row++ {
		for _, col := range []string{"a", "b", "c", "d", "e", "f"} {
			if t, err := conn.Exec(ctx, "INSERT INTO seats (col, row) VALUES ($1, $2);", col, row); err != nil {
				log.Error().Err(err).Msg("failed to insert row")
			} else {
				log.Debug().Msg(t.String())
			}
		}
	}

	show(ctx, conn)

	startTime := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(6 * rowCount)
	for u := 1; u <= 6*rowCount; u++ {
		go func(uid int, wg *sync.WaitGroup) {
			defer wg.Done()
			rows, err := conn.Query(ctx, `
				UPDATE seats
				SET user_id = $1
				FROM (
					SELECT id
					FROM seats
					WHERE user_id = 0
					LIMIT 1
					FOR UPDATE SKIP LOCKED
				) AS subquery
				WHERE seats.id = subquery.id
				RETURNING seats.id as id, row, col, user_id;`,
				uid,
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to update")
				return
			}
			seats, err := pgx.CollectRows(rows, pgx.RowToStructByName[Seat])
			if err != nil {
				log.Error().Err(err).Msg("failed to collect rows")
			} else {
				log.Debug().Interface("seat", seats).Msg("row updated")
			}
		}(u, wg)
	}

	wg.Wait()
	endTime := time.Now()

	log.Info().
		Str("delta", endTime.Sub(startTime).String()).
		Msg("time taken")

	show(ctx, conn)

	if _, err := conn.Exec(ctx, `DROP TABLE seats;`); err != nil {
		log.Error().Err(err).Msg("failed to drop table")
	}
}
