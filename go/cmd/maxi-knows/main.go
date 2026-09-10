package main

import (
	"encoding/json"
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

	http.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Login")

	})


	http.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request){
		q := r.URL.Query().Get("q")

		language := r.URL.Query().Get("language")

		//Replace with actual result from db later:
		_= q
		_= language

		if !r.URL.Query().Has("q"){
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 422,
				"message": "q is required",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			//replace with actual matches later:
			"data": []interface{}{},
		})
	})

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
