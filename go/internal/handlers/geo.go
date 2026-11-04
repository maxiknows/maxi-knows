package handlers

import (
	"errors"
	"log"
	"net/http"

	"maxi-knows/internal/geo"
)

func WeatherHandler(svc *geo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		lat, lon, ok := geo.ParseCoords(q.Get("lat"), q.Get("lon"))
		if !ok {
			http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
			return
		}

		body, hit, err := svc.Weather(lat, lon)
		if err != nil {
			log.Printf("Weather fetch failed: %v", err)
			http.Error(w, "Failed to fetch weather", http.StatusBadGateway)
			return
		}

		writeCachedJSON(w, body, hit)
	}
}

func LocationHandler(svc *geo.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		lat, lon, ok := geo.ParseCoords(q.Get("lat"), q.Get("lon"))
		if !ok {
			http.Error(w, "valid lat and lon are required", http.StatusBadRequest)
			return
		}

		body, hit, err := svc.Location(lat, lon)
		if errors.Is(err, geo.ErrLocationNotFound) {
			http.Error(w, "Location not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Location lookup failed: %v", err)
			http.Error(w, "Failed to look up location", http.StatusBadGateway)
			return
		}

		writeCachedJSON(w, body, hit)
	}
}

func writeCachedJSON(w http.ResponseWriter, body []byte, hit bool) {
	w.Header().Set("Content-Type", "application/json")
	if hit {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}
	_, _ = w.Write(body)
}
