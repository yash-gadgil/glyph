package handlers

import "net/http"

func GetAccount(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your account"))
}

func GetFunds(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your funds"))
}

func GetProfile(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your profile"))
}

func GetTrades(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your trades"))
}
