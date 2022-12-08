package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	// "golang.org/x/text/message"
)

var DB *pgxpool.Pool
// var TELEGRAM="https://api.telegram.org/bot$code/sendMessage"
type Todos struct {
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
	User_id   int    `json:"user_id"`
}
type Message struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	//!load env variables
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
	// Init the mux router
	router := mux.NewRouter()
	// Get all todos
	router.HandleFunc("/todos/", getAllTodos).Methods("POST")
	// Create a todo
	router.HandleFunc("/todos/create", createTodo).Methods("POST")
	// Delete a specific todo by the id
	router.HandleFunc("/todos/delete", deleteTodo).Methods("DELETE")
	// Update  todo
	router.HandleFunc("/todos/edit", updateTodo).Methods("PATCH")
	// serve the app
	go checkForMessages()
	// fmt.Scanln()
	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8000", router))

}

// !repititive message checker
func checkForMessages() {
	ticker := time.NewTicker(3 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			//1048346950
			type Todo struct {
				Remind_at   time.Time `json:"remind_at"`
				Is_reminded bool      `json:"is_reminded"`
				User_id     int       `json:"user_id"`
			}
			type User struct {
				Id                 int    `json:"id"`
				Telegram_chat_id   string `json:"telegram_chat_id"`
				Telegram_bot_token string `json:"telegram_bot_token"`
			}
			var p Todo
			var u User
			r, err := DB.Query(context.Background(), "select remind_at,is_reminded,user_id from todo")
			if err != nil {
				panic(err)
			}
			defer r.Close()
			for r.Next() {
				err := r.Scan(&p.Remind_at, &p.Is_reminded, &p.User_id)
				if err != nil {
					panic(err)
				}
				rows, err := DB.Query(context.Background(), "select id,telegram_bot_token,telegram_chat_id from users")
				if err != nil {
					fmt.Println(err.Error())
					return
					// panic(err)
				}
				defer rows.Close()
				for rows.Next() {
					err := rows.Scan(&u.Id, &u.Telegram_bot_token, &u.Telegram_chat_id)
					if err != nil {
						fmt.Println(err.Error())
						return
					}
					if p.User_id == u.Id {
						if p.Is_reminded {
							currentTime := time.Now()
							oldTime := p.Remind_at
							diff := currentTime.Sub(oldTime)
							fmt.Printf("Seconds: %f\n", diff.Seconds())
							if diff.Seconds() == 0 {
								i, err := strconv.ParseInt(u.Telegram_chat_id, 10, 64)
								if err != nil {
									fmt.Println(err.Error())
									return
								}
								message := Message{ChatID: i, Text: "salam"}
								response := "https://api.telegram.org/bot" + u.Telegram_bot_token + "/sendMessage"
								// response := fmt.Sprintf("https://api.telegram.org/bot%d/sendMessage", u.Telegram_bot_token)
								SendMessage(response, &message)
							}

						}
					}
				}
				err = rows.Err()
				if err != nil {
					panic(err)
				}
			}
			err = r.Err()
			if err != nil {
				panic(err)
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}

}

// !initialize a database
func initDb() (err error) {
	conf, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}
	conf.MaxConns = 5
	DB, err = pgxpool.ConnectConfig(context.Background(), conf)
	return err
}

// !send message to a telegram bot
func SendMessage(url string, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}
	return nil
}

// !get all todos
func getAllTodos(w http.ResponseWriter, r *http.Request) {
	type Todo struct {
		User_id int `json:"user_id"`
	}
	var p Todo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	str := "select content,completed,user_id from todo WHERE user_id=$1"
	id := strconv.Itoa(p.User_id)
	rows, err := DB.Query(context.Background(), str, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer rows.Close()
	var rowSlice []Todos
	for rows.Next() {
		var r Todos
		err := rows.Scan(&r.Content, &r.Completed, &r.User_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rowSlice = append(rowSlice, Todos{Content: r.Content, Completed: r.Completed, User_id: r.User_id})
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rowSlice)
}

// !create todo
func createTodo(w http.ResponseWriter, r *http.Request) {
	type Todo struct {
		Content string `json:"content"`
		User_id int    `json:"user_id"`
	}
	var p Todo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := DB.BeginTx(context.Background(), pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var lim int
	str := "select lim from users where id = $1"
	er := tx.QueryRow(context.Background(), str, p.User_id).Scan(&lim)
	if er != nil && err != sql.ErrNoRows {
		fmt.Println(er.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var count int
	coun := "select count(0) from todo where user_id = $1"
	er = tx.QueryRow(context.Background(), coun, p.User_id).Scan(&count)
	if er != nil {
		fmt.Println(er.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if lim > count {
		var lastInsertId int
		coun := "select count(0) from todo"
		er = tx.QueryRow(context.Background(), coun).Scan(&lastInsertId)
		if er != nil {
			fmt.Println(er.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res, err := tx.Exec(context.Background(), "INSERT INTO todo (content,user_id) VALUES($1, $2) RETURNING id", p.Content, p.User_id)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(res)
		err = tx.Commit(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}
		type Id struct {
			Id int `json:"id"`
		}
		resp := Id{lastInsertId + 1}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Limit exceeded!"))
	fmt.Println("Limit kecildi")
}

// !update todo
func updateTodo(w http.ResponseWriter, r *http.Request) {
	type Todo struct {
		Completed bool
		Id        int
	}
	var p Todo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	str4 := strconv.Itoa(p.Id)
	str2 := strconv.FormatBool(p.Completed)
	_, err = DB.Exec(context.Background(), "UPDATE todo SET completed = $1 WHERE id=$2 ", str2, str4)
	if err != nil {
		log.Fatal(err)
		return
	}
	type Success struct {
		Success string `json:"success"`
	}
	resp := Success{"true"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

// !delete todo
func deleteTodo(w http.ResponseWriter, r *http.Request) {
	type Todo struct {
		Id int `json:"id"`
	}
	var p Todo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	str := "DELETE FROM todo WHERE id = $1"
	_, err = DB.Exec(context.Background(), str, p.Id)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	type Success struct {
		Success string `json:"success"`
	}
	resp := Success{"true"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
