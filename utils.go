package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func newReadCloserFromURL(u url.URL) (io.ReadCloser, error) {
	switch u.Scheme {
	case "http":
		log.Printf("Requesting URL: %s", u.String())
		request, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		return response.Body, nil
	case "file":
		absPath := u.Path
		if absPath == "" && u.Opaque != "" {
			p, err := filepath.Abs(u.Opaque)
			if err != nil {
				return nil, err
			}
			absPath = p
		}
		log.Printf("Opening file URL at absolute path: %s", absPath)
		return os.Open(absPath)
	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}
