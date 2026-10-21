package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main(){
	mux := http.NewServeMux()

	mux .HandleFunc("GET /health" ,func(w http.ResponseWriter , r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":"ok",
		})
	})

	log.Println("LinkForge listening on :8080")
	log.Fatal(http.ListenAndServe(":8080",mux))
}