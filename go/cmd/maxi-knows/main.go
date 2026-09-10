package main

import (
	"fmt"
	"net/http"
)

func main(){
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "WhoKnows")
	})


	http.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Register")
	})

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
