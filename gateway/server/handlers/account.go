package handlers

import "net/http"

func (cfg *Config) GetAccount(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your account"))
}

func (cfg *Config) GetFunds(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your funds"))
}

func (cfg *Config) GetProfile(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your profile"))
}

func (cfg *Config) GetTrades(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your trades"))
}
