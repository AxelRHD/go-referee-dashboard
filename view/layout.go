package view

import (
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

func BasePage(pageTitle string, content ...g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    pageTitle + " — Referee Dashboard",
		Language: "de",
		Head: []g.Node{
			h.Meta(g.Attr("charset", "utf-8")),
			h.Meta(g.Attr("name", "viewport"), g.Attr("content", "width=device-width, initial-scale=1")),
			h.Link(g.Attr("rel", "icon"), g.Attr("type", "image/png"), g.Attr("href", "/static/pfeife.png")),
			// Bootstrap 5.3.3
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css")),
			// Bootstrap Icons
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css")),
			// Nord Theme
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "/static/css/nord.css")),
			// Theme init script
			h.Script(g.Raw(`
				(function() {
					const theme = localStorage.getItem('theme') || 'dark';
					document.documentElement.setAttribute('data-bs-theme', theme);
				})();
			`)),
		},
		Body: []g.Node{
			navbar(),
			h.Div(h.Class("container mt-4"),
				g.Group(content),
			),
			// Bootstrap JS
			h.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js")),
			// HTMX
			h.Script(g.Attr("src", "https://unpkg.com/htmx.org@2")),
			// Plotly.js
			h.Script(g.Attr("src", "https://cdn.plot.ly/plotly-3.0.1.min.js")),
			// Alpine.js
			h.Script(g.Attr("defer", ""), g.Attr("src", "https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js")),
		},
	})
}

func navbar() g.Node {
	type navLink struct {
		href  string
		label string
	}
	links := []navLink{
		{"/dashboard/", "Dashboard"},
		{"/games", "Spiele"},
		{"/teams", "Teams"},
		{"/leagues", "Ligen"},
		{"/venues", "Spielorte"},
		{"/data", "Datenverwaltung"},
	}

	var navItems []g.Node
	for _, l := range links {
		navItems = append(navItems, h.Li(h.Class("nav-item"),
			h.A(h.Class("nav-link"), g.Attr("href", l.href), g.Text(l.label)),
		))
	}

	return h.Nav(h.Class("navbar navbar-expand-lg bg-body-tertiary"),
		h.Div(h.Class("container"),
			h.A(h.Class("navbar-brand"), g.Attr("href", "/"),
				h.Img(g.Attr("src", "/static/pfeife.png"), g.Attr("alt", ""), g.Attr("width", "30"), g.Attr("height", "30"), h.Class("d-inline-block align-text-top me-2")),
				g.Text("Referee Dashboard"),
			),
			h.Button(h.Class("navbar-toggler"), g.Attr("type", "button"),
				g.Attr("data-bs-toggle", "collapse"), g.Attr("data-bs-target", "#navbarNav"),
				h.Span(h.Class("navbar-toggler-icon")),
			),
			h.Div(h.Class("collapse navbar-collapse"), g.Attr("id", "navbarNav"),
				h.Ul(h.Class("navbar-nav me-auto"),
					g.Group(navItems),
				),
				themeToggle(),
			),
		),
	)
}

func themeToggle() g.Node {
	return h.Div(
		g.Attr("x-data", `{
			theme: localStorage.getItem('theme') || 'dark',
			toggle() {
				this.theme = this.theme === 'dark' ? 'light' : 'dark';
				localStorage.setItem('theme', this.theme);
				document.documentElement.setAttribute('data-bs-theme', this.theme);
				window.dispatchEvent(new CustomEvent('theme-changed'));
			}
		}`),
		h.Button(h.Class("btn btn-outline-secondary btn-sm"),
			g.Attr("@click", "toggle()"),
			h.I(h.Class("bi"), g.Attr(":class", "theme === 'dark' ? 'bi-sun' : 'bi-moon'")),
		),
	)
}
