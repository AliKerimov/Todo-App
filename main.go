package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	// "github.com/jackc/pgx//pgxpool"
	"log"
	"net/http"
	"os"
	"strconv"
	// "text/template"

	_ "github.com/lib/pq"
)

var DB *pgxpool.Pool

//var tx *sql.Tx

type Todos struct {
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
	User_id   int    `json:"user_id"`
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
	// Route handles & endpoints
	// Get all todos
	router.HandleFunc("/todos/", getAllTodos).Methods("POST")
	// Create a todo
	router.HandleFunc("/todos/create", createTodo).Methods("POST")
	// Delete a specific todo by the id
	router.HandleFunc("/todos/delete", deleteTodo).Methods("DELETE")
	// Update  todo
	router.HandleFunc("/todos/edit", updateTodo).Methods("PATCH")
	// serve the app
	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8000", router))

}
func initDb() (err error) {
	conf, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}

	conf.MaxConns = 5

	DB, err = pgxpool.ConnectConfig(context.Background(), conf)
	return err
}

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
	//fmt.Println(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//fmt.Printf(err)
		return
	}

	var lim int
	str := "select lim from users where id = $1"
	er := tx.QueryRow(context.Background(), str, p.User_id).Scan(&lim)
	if er != nil && err != sql.ErrNoRows {
		//tx.Rollback(context.Background())
		fmt.Println(er.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println("skjskj")

	var count int
	coun := "select count(0) from todo where user_id = $1"
	er = tx.QueryRow(context.Background(), coun, p.User_id).Scan(&count)
	if er != nil {
		//tx.Rollback(context.Background())
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
		// s, err := DB.Exec(context.Background(), "INSERT INTO todo (content,user_id) VALUES($1, $2)", p.Content, p.User_id)
		res, err := tx.Exec(context.Background(), "INSERT INTO todo (content,user_id) VALUES($1, $2) RETURNING id", p.Content, p.User_id)
		if err != nil {
			//tx.Rollback(context.Background())
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
		//lastInsertId = lastInsertId + 1
		resp := Id{lastInsertId + 1}
		//resp := 2
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Limit exceeded!"))
	fmt.Println("Limit kecildi")
}
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
	// l:=p.Id
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
