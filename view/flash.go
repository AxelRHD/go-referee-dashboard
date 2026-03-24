package view

import (
	"net/http"
	"net/url"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const flashCookieName = "flash"

func SetFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    url.QueryEscape(msg),
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
	})
}

func GetFlash(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(flashCookieName)
	if err != nil {
		return ""
	}
	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   flashCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	msg, _ := url.QueryUnescape(c.Value)
	return msg
}

func FlashAlert(w http.ResponseWriter, r *http.Request) g.Node {
	msg := GetFlash(w, r)
	if msg == "" {
		return nil
	}
	return h.Div(h.Class("alert alert-success alert-dismissible fade show"),
		g.Text(msg),
		h.Button(h.Class("btn-close"), g.Attr("type", "button"), g.Attr("data-bs-dismiss", "alert")),
	)
}
