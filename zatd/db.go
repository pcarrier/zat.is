package main

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func insertUrlWhenAbsent(path string, url string, ctx context.Context) (err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, "INSERT INTO shorts(short,url) VALUES ($1, $2) ON CONFLICT DO NOTHING", path, url)
	return
}

func lookUp(path string, ctx context.Context) (target string, err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()
	err = conn.QueryRow(ctx, "SELECT url FROM shorts WHERE short = $1", path).Scan(&target)
	if err == pgx.ErrNoRows {
		err = nil
	}
	return
}
