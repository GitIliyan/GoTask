package main

import (
	"fmt"
	"net/http"

	"github.com/GitIliyan/GoTask/internal"
)

func init() {
	fmt.Println("initializing the map")
	internal.InitializeMap()
}
func main() {
	http.HandleFunc("/extractNumber", internal.ExtractHandler)
	http.HandleFunc("/validateExpression", internal.ValidateHandler)
	http.HandleFunc("/errorFrequencies", internal.GetErrorFrequenciesHandler)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
