package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/price"
)

var storeItems = map[int]map[string]interface{}{
	1: {"priceInCents": 10000, "name": "Pivo"},
	2: {"priceInCents": 20000, "name": "glasses"},
}

func main() {
	stripe.PrivateKey = os.Getenv("STRIPE_PRIVATE_KEY")

	r := mux.NewRouter()

	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5500"},
	}).Handler)

	r.HandleFunc("/checkout/{itemId}", handleCheckout).Methods("POST")

	port := "3000"
	log.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	item, exists := storeItems[itemID]
	if !exists {
		http.NotFound(w, r)
		return
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    price.NewParams().SetUnitAmount(int64(item["priceInCents"].(float64))).SetCurrency("usd").SetProductData(&stripe.PriceProductDataParams{Name: stripe.String(item["name"].(string))}),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String("payment"),
		SuccessURL: stripe.String("http://localhost:5500/success.html"),
		CancelURL:  stripe.String("http://localhost:5500/cancel.html"),
	}

	session, err := stripe.CheckoutSessions.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"id": session.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}