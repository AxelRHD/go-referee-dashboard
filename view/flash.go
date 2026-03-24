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
	SetFlashes(w, []string{msg})
}

func SetFlashes(w http.ResponseWriter, messages []string) {
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
	http.SetCookie(w, &http.Cookie{
		Name:   flashCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	decoded, _ := url.QueryUnescape(c.Value)
	var messages []string
	if err := json.Unmarshal([]byte(decoded), &messages); err != nil {
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
