package scraper

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Cookie represents a cookie, it copies parts of the http.Cookie struct but changes
// the JSON marshaling to exclude empty fields.
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`

	Expires *time.Time `json:"expires,omitempty"`
}

// Cookies returns the current cookies.
func (s *Scraper) Cookies() []Cookie {
	httpCookies := s.cookies.Cookies(s.URL)
	cookies := make([]Cookie, 0, len(httpCookies))

	for _, c := range httpCookies {
		cookie := Cookie{
			Name:  c.Name,
			Value: c.Value,
		}
		if !c.Expires.IsZero() {
			cookie.Expires = &c.Expires
		}
		cookies = append(cookies, cookie)
	}

	return cookies
}

func createCookieJar(u *url.URL, cookies []Cookie) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	httpCookies := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		h := &http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		}
		if c.Expires != nil {
			h.Expires = *c.Expires
		}
		httpCookies = append(httpCookies, h)
	}

	jar.SetCookies(u, httpCookies)
	return jar, nil
}
