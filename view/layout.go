package view

import (
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

func BasePage(pageTitle string, content ...g.Node) g.Node {
	return basePage(pageTitle, "container", content...)
}

func BasePageRaw(pageTitle string, content ...g.Node) g.Node {
	return basePage(pageTitle, "", content...)
}

func basePage(pageTitle, container string, content ...g.Node) g.Node {
	var contentWrapper g.Node
	if container != "" {
		contentWrapper = h.Div(h.Class(container+" mt-4"), g.Group(content))
	} else {
		contentWrapper = h.Div(h.Class("mt-4"), g.Group(content))
	}

	return c.HTML5(c.HTML5Props{
		Title:    pageTitle + " — Referee Dashboard",
		Language: "de",
		Head: []g.Node{
			h.Meta(g.Attr("charset", "utf-8")),
			h.Meta(g.Attr("name", "viewport"), g.Attr("content", "width=device-width, initial-scale=1")),
			h.Link(g.Attr("rel", "icon"), g.Attr("type", "image/png"), g.Attr("href", "/static/pfeife.png")),
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css")),
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css")),
			h.Link(g.Attr("rel", "stylesheet"), g.Attr("href", "/static/css/nord.css")),
			h.Script(g.Raw(`(function(){var t=localStorage.getItem('theme')||'dark';document.documentElement.setAttribute('data-bs-theme',t)})();`)),
		},
		Body: []g.Node{
			navbar(),
			contentWrapper,
			h.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js")),
			h.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/htmx.org@2/dist/htmx.min.js")),
			h.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/plotly.js-dist-min@2/plotly.min.js")),
			h.Script(g.Raw(`
document.addEventListener('alpine:init',()=>{
    Alpine.data('themeToggle',()=>({
        dark:document.documentElement.getAttribute('data-bs-theme')==='dark',
        toggle(){
            this.dark=!this.dark;
            var theme=this.dark?'dark':'light';
            document.documentElement.setAttribute('data-bs-theme',theme);
            localStorage.setItem('theme',theme);
            window.dispatchEvent(new Event('theme-changed'));
        }
    }));
});`)),
			h.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js")),
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
		navItems = append(navItems,
			h.A(h.Class("nav-link"), g.Attr("href", l.href), g.Text(l.label)),
		)
	}

	return h.Nav(h.Class("navbar navbar-expand-lg sticky-top"),
		h.Div(h.Class("container"),
			h.A(h.Class("navbar-brand d-flex align-items-center"), g.Attr("href", "/"),
				h.Img(g.Attr("src", "/static/pfeife.png"), g.Attr("alt", ""), g.Attr("width", "28"), g.Attr("height", "28"), h.Class("me-2")),
				g.Text("Referee Dashboard"),
			),
			h.Div(h.Class("d-flex align-items-center"),
				h.Div(h.Class("navbar-nav me-auto"),
					g.Group(navItems),
				),
				themeToggle(),
			),
		),
	)
}

func themeToggle() g.Node {
	return h.Div(g.Attr("x-data", "themeToggle"),
		h.Button(h.Class("theme-toggle"),
			g.Attr("@click", "toggle()"),
			g.Attr("title", "Theme wechseln"),
			h.I(h.Class("bi bi-sun"), g.Attr("x-show", "dark")),
			h.I(h.Class("bi bi-moon"), g.Attr("x-show", "!dark")),
		),
	)
}
