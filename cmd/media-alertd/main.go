package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	mediaalert "github.com/infrai-examples/media-asset-sms"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	notifier := mediaalert.DeliveryNotifier{SMS: mediaalert.SMSClient{APIKey: apiKey, MaxRetries: 3}}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /asset-events", func(response http.ResponseWriter, request *http.Request) {
		var event mediaalert.AssetEvent
		decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&event); err != nil {
			http.Error(response, "invalid asset event", http.StatusBadRequest)
			return
		}
		result, sent, err := notifier.Handle(request.Context(), event)
		if err != nil {
			var apiErr *mediaalert.APIError
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
				http.Error(response, "asset event was rejected", apiErr.Status)
				return
			}
			log.Printf("asset %q: %v", event.AssetID, err)
			http.Error(response, "delivery failed", http.StatusBadGateway)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		if !sent {
			json.NewEncoder(response).Encode(map[string]any{"sent": false})
			return
		}
		json.NewEncoder(response).Encode(map[string]any{"sent": true, "data": result.Data, "metadata": result.Metadata})
	})

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("media-alertd listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
