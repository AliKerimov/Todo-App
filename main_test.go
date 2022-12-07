package main

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v4"
	// "os"
	"fmt"
	"log"
	"net/http"
	"testing"
	"github.com/joho/godotenv"
	"time"
)
var D *pgx.Conn

func TestCreateTodo(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	err = initDb()
	if err != nil {
		log.Fatal(err)
		return
	}
	for i := 0; i < 100; i++ {
		go test(100)
	}
	time.Sleep(time.Millisecond*1000)
	var counter int
	c := "select count(0) from todo where user_id = $1"
	er := D.QueryRow(context.Background(), c, 3).Scan(&counter)
	if er != nil {
		log.Fatal(er)
		return
	}
	fmt.Scanln(counter)
}
func test(count int) {
	for i := 0; i < count; i++ {
		posturl := "http://localhost:8000/todos/create"
		body := []byte(`{
		"content": "salam",
		"userId": 3
	}`)
		req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
		http.DefaultClient.Do(req)
	}
}
