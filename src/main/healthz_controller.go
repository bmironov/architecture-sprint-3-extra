package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello"))
}

func livez(w http.ResponseWriter, req *http.Request) {
	fmt.Println("/livez endpoint called")
	w.WriteHeader(http.StatusOK)
}

func readyz(w http.ResponseWriter, req *http.Request) {
	fmt.Println("/readyz endpoint called")

	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	dbFlag = true

	w.WriteHeader(http.StatusOK)
}
