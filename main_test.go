package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var D *pgxpool.Pool

func initD() (err error) {
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	config.MaxConns = 5
	DB, err = pgxpool.ConnectConfig(context.Background(), config)
	fmt.Println("connected.....")
	return err
}
func TestCreateTodo(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	err = initD()
	if err != nil {
		fmt.Println("hi")
		log.Fatal(err)
		return
	}
	for i := 0; i < 100; i++ {
		go test(100)
	}
	time.Sleep(time.Second * 10)
	var counter int
	c := "select count(0) from todo where user_id = $1"
	er := D.QueryRow(context.Background(), c, 7).Scan(&counter)
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
		"user_id": 7
	}`)
		req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
			//return
		}
		//client := &http.Client{}
		//res, err := client.Do(req)
		//if err != nil {
		//	panic(err)
		//}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
		//log.Fatal(res)
	}
}
