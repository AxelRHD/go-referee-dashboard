package view

import (
	"encoding/json"
	"net/http"
	"net/url"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const flashCookieName = "flash"

func SetFlash(w http.ResponseWriter, msg string) {
	// Read existing messages from Set-Cookie headers
	var messages []string
	for _, c := range w.Header()["Set-Cookie"] {
		if len(c) > len(flashCookieName)+1 && c[:len(flashCookieName)+1] == flashCookieName+"=" {
			val := c[len(flashCookieName)+1:]
			if idx := indexOf(val, ';'); idx >= 0 {
				val = val[:idx]
			}
			decoded, _ := url.QueryUnescape(val)
			_ = json.Unmarshal([]byte(decoded), &messages)
			break
		}
	}
	messages = append(messages, msg)
	data, _ := json.Marshal(messages)
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    url.QueryEscape(string(data)),
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
	})
}

func GetFlashes(w http.ResponseWriter, r *http.Request) []string {
	c, err := r.Cookie(flashCookieName)
	if err != nil {
		return nil
	}
	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   flashCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	decoded, _ := url.QueryUnescape(c.Value)
	var messages []string
	if err := json.Unmarshal([]byte(decoded), &messages); err != nil {
		// Fallback: single string (old format)
		if decoded != "" {
			return []string{decoded}
		}
		return nil
	}
	return messages
}

func FlashAlert(w http.ResponseWriter, r *http.Request) g.Node {
	messages := GetFlashes(w, r)
	if len(messages) == 0 {
		return nil
	}
	var alerts []g.Node
	for _, msg := range messages {
		alerts = append(alerts, h.Div(h.Class("alert alert-success alert-dismissible fade show"),
			g.Text(msg),
			h.Button(h.Class("btn-close"), g.Attr("type", "button"), g.Attr("data-bs-dismiss", "alert")),
		))
	}
	return g.Group(alerts)
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
