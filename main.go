package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type tpl struct {
	Posts []*Post
	Post  *Post
}

type Post struct {
	Id      int
	Title   string
	Author  sql.NullString
	Text    string
	Updated sql.NullString
}

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	posts := []*Post{}

	rows, err := h.DB.Query("SELECT id, title, author, text, updated FROM posts")
	check(err)
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.Id, &post.Title, &post.Author, &post.Text, &post.Updated)
		check(err)
		posts = append(posts, post)
	}
	rows.Close()

	err = h.Tmpl.ExecuteTemplate(w, "index.html", tpl{
		Posts: posts,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	author := r.FormValue("author")
	text := r.FormValue("text")

	if title == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `title` is requeired! </p>")
		return
	}
	if author == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `author` is requeired! </p>")
		return
	}
	if text == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `text` is requeired! </p>")
		return
	}

	result, err := h.DB.Exec("INSERT INTO posts (`title`, `author`, `text`) VALUES (?, ?, ?)", title, author, text)
	check(err)
	newid, _ := result.LastInsertId()
	addcnt, _ := result.RowsAffected()
	fmt.Printf("Created!\n\tLast id: %v\n\tAdded rows: %v", newid, addcnt)
	fmt.Println(result.RowsAffected())

	http.Redirect(w, r, "/posts", http.StatusFound)
}

func (h *Handler) AddPost(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "add.html", nil)
	check(err)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TRY DELETE")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	res, err := h.DB.Exec("DELETE FROM posts WHERE id = ?", id)
	check(err)

	affected, err := res.RowsAffected()
	check(err)

	fmt.Println("DELETE by [" + r.UserAgent() + "]")
	fmt.Printf("\tid: %v\n\taffected: %v\n", id, affected)

	fmt.Println("DELETED ", id)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	rows, err := h.DB.Query("SELECT id, title, author, text, updated FROM posts WHERE id = ?", id)
	check(err)
	post := &Post{}
	for rows.Next() {
		err = rows.Scan(&post.Id, &post.Title, &post.Author, &post.Text, &post.Updated)
		check(err)
	}
	rows.Close()

	h.Tmpl.ExecuteTemplate(w, "edit.html", post)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	title := r.FormValue("title")
	text := r.FormValue("text")
	updated := r.FormValue("updated")
	if title == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `title` is requeired! </p>")
		return
	}
	if updated == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `updated` is requeired! </p>")
		return
	}
	if text == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `text` is requeired! </p>")
		return
	}

	res, err := h.DB.Exec("UPDATE posts SET"+
		"`title` = ?,"+
		"`text` = ?,"+
		"`updated` = ?"+
		" WHERE id = ?",
		title,
		text,
		updated,
		id,
	)
	check(err)

	affected, err := res.RowsAffected()
	check(err)
	fmt.Println("UPDATED BY [" + r.UserAgent() + "]")
	fmt.Printf("\tid: %v\n\taffected: %v\n", id, affected)

	http.Redirect(w, r, "/posts", http.StatusFound)
}

func main() {
	fmt.Println("Connecting to database...")
	dsn := "root:pass@tcp(localhost)/database?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(10)
	check(err)

	fmt.Println("Ping!")
	err = db.Ping()
	check(err)
	fmt.Println("     Pong!")

	handlers := &Handler{
		DB:   db,
		Tmpl: template.Must(template.ParseGlob("templates/*")),
	}

	r := mux.NewRouter()
	r.HandleFunc("/posts", handlers.Index).Methods("GET")
	r.HandleFunc("/posts/add", handlers.AddPost).Methods("GET")
	r.HandleFunc("/posts/add", handlers.Add).Methods("POST")
	r.HandleFunc("/posts/edit/{id}", handlers.Edit).Methods("GET")
	r.HandleFunc("/posts/edit/{id}", handlers.Update).Methods("POST")
	r.HandleFunc("/posts/delete/{id}", handlers.Delete).Methods("DELETE")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)
}
