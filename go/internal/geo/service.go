package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrLocationNotFound = errors.New("location not found")

type Service struct {
	weather  *ttlCache
	location *ttlCache
}

func NewService() *Service {
	return &Service{
		// Forecasts go stale quickly, place names practically never do.
		weather:  newTTLCache(15 * time.Minute),
		location: newTTLCache(24 * time.Hour),
	}
}

// cached returns the cached value or calls fetch and stores the result.
// The bool reports whether it was a cache hit.
func cached(c *ttlCache, key string, fetch func() ([]byte, error)) ([]byte, bool, error) {
	if data, found := c.get(key); found {
		return data, true, nil
	}
	data, err := fetch()
	if err != nil {
		return nil, false, err
	}
	c.set(key, data)
	return data, false, nil
}

// Weather returns the raw Open-Meteo JSON for already-validated coordinates.
func (s *Service) Weather(lat, lon string) ([]byte, bool, error) {
	return cached(s.weather, lat+","+lon, func() ([]byte, error) {
		return fetchBody(fmt.Sprintf(
			"https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,weather_code&daily=temperature_2m_max,temperature_2m_min,weather_code&timezone=auto",
			lat, lon,
		))
	})
}

// Location returns {"name": "Copenhagen, Denmark"} as JSON.
// Returns ErrLocationNotFound if Nominatim has no usable name.
func (s *Service) Location(lat, lon string) ([]byte, bool, error) {
	return cached(s.location, lat+","+lon, func() ([]byte, error) {
		// zoom=10 asks for city-level detail rather than street-level.
		body, err := fetchBody(fmt.Sprintf(
			"https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=%s&lon=%s&zoom=10&accept-language=en",
			lat, lon,
		))
		if err != nil {
			return nil, err
		}

		var geo struct {
			Address struct {
				City         string `json:"city"`
				Town         string `json:"town"`
				Village      string `json:"village"`
				Municipality string `json:"municipality"`
				County       string `json:"county"`
				Country      string `json:"country"`
			} `json:"address"`
		}
		if err := json.Unmarshal(body, &geo); err != nil {
			return nil, fmt.Errorf("decode nominatim response: %w", err)
		}

		// Nominatim uses different keys depending on place size; take the first filled one.
		name := ""
		for _, part := range []string{
			geo.Address.City, geo.Address.Town, geo.Address.Village,
			geo.Address.Municipality, geo.Address.County,
		} {
			if part != "" {
				name = part
				break
			}
		}
		if geo.Address.Country != "" {
			if name != "" {
				name += ", "
			}
			name += geo.Address.Country
		}
		if name == "" {
			return nil, ErrLocationNotFound
		}

		return json.Marshal(map[string]string{"name": name})
	})
}
