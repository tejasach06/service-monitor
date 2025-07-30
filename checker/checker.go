package checker

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func sanitizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

func CheckService(host string, port int, path string, timeout time.Duration) (bool, string) {
	path = sanitizePath(path)

	url := fmt.Sprintf("http://%s:%d%s", host, port, path)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true, "http"
		}
	}

	// HTTPS attempt
	url = fmt.Sprintf("https://%s:%d%s", host, port, path)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	resp, err = (&http.Client{Timeout: timeout, Transport: tr}).Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true, "https"
		}
	}
	return false, ""
}
