package geo

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func fetchBody(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Nominatim's usage policy requires an identifying User-Agent.
	req.Header.Set("User-Agent", "whoknows-search/1.0 (school project)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
