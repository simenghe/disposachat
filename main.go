package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func main() {
	const PORT = ":3000"
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("Hello World"))
	})
	fmt.Println("Running on port " + PORT)
	err := http.ListenAndServe(PORT, r)
	fmt.Println(err)
}
