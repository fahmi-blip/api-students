package main

import (
	"context"
	"fmt"
	"log"

	"api-students/config"
	"api-students/database"
)

func main() {
	config.LoadEnv()
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("=== USERS ===")
	rows, _ := pool.Query(ctx, "SELECT id, username, email FROM users ORDER BY id DESC LIMIT 5")
	for rows.Next() {
		var id int
		var u, e string
		rows.Scan(&id, &u, &e)
		fmt.Printf("User: id=%d, username=%s, email=%s\n", id, u, e)
	}
	rows.Close()

	fmt.Println("=== ENROLLMENTS ===")
	rows, _ = pool.Query(ctx, "SELECT id, course_id, user_id, created_at FROM enrollments ORDER BY id DESC LIMIT 10")
	for rows.Next() {
		var id, cid, uid int
		var t any
		rows.Scan(&id, &cid, &uid, &t)
		fmt.Printf("Enrollment: id=%d, course_id=%d, user_id=%d, time=%v\n", id, cid, uid, t)
	}
	rows.Close()

	fmt.Println("=== OUTBOX ===")
	rows, _ = pool.Query(ctx, "SELECT id, event_type, published_at IS NOT NULL, payload FROM outbox ORDER BY id DESC LIMIT 10")
	for rows.Next() {
		var id int
		var et string
		var pub bool
		var p []byte
		rows.Scan(&id, &et, &pub, &p)
		fmt.Printf("Outbox: id=%d, type=%s, published=%v, payload=%s\n", id, et, pub, string(p))
	}
	rows.Close()

	fmt.Println("=== NOTIFICATIONS ===")
	rows, _ = pool.Query(ctx, "SELECT id, outbox_id, user_id, subject, sent_at FROM notifications ORDER BY id DESC LIMIT 10")
	for rows.Next() {
		var id, oid, uid int
		var s string
		var t any
		rows.Scan(&id, &oid, &uid, &s, &t)
		fmt.Printf("Notification: id=%d, outbox_id=%d, user_id=%d, subject=%s, sent_at=%v\n", id, oid, uid, s, t)
	}
	rows.Close()
}

