package handlers

import "net/http"

func (cfg *Config) GetPortfolio(w http.ResponseWriter, r *http.Request) {

	/* serverAddr := "http://localhost:3001"
	conn, err := grpc.NewClient(serverAddr)
	if err != nil {
		fmt.Println("Failed to start gRPC Client")
	}
	defer conn.Close() */

	w.Write([]byte("your portfolio"))
}

func (cfg *Config) GetHoldings(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("your holdings"))
}

func (cfg *Config) GetPositions(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("your positions"))
}
