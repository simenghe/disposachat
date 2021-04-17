package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
)

var Tbl map[string]int

type response1 struct {
	Token  string
	Value  string
	Expiry time.Duration
}

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func StartSession(key string) {
	Tbl[key] = 1
	time.Sleep(10 * time.Second)
	defer fmt.Printf("Session %s ended\n", key)
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d commenced\n", id)
	start := time.Now()
	time.Sleep(time.Second * 2)
	end := start.Add(time.Duration(time.Second * 2))
	fmt.Println(start)
	fmt.Println(end)
	fmt.Printf("Worker %d finished\n", id)
}

func Ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	fmt.Println(s)
	return err
}

func Setex(c redis.Conn, key string, value string, timeMinutes int) (string, error) {
	s, err := redis.String(c.Do("SETEX", key, timeMinutes*60, value))
	return s, err
}

func main() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	fmt.Println(conn)
	defer conn.Close()

	err = Ping(conn)
	if err != nil {
		panic(err)
	}

	const PORT = ":3000"
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	r.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		token := randToken()
		duration := 3
		res, err := Setex(conn, token, "Good Morning", duration)
		if err != nil {
			panic(err)
		}
		result := &response1{token, res, time.Duration(duration)}
		data, _ := json.Marshal(result)
		w.Write(data)
	})
	fmt.Println("Running on port " + PORT)
	http.ListenAndServe(PORT, r)
}
