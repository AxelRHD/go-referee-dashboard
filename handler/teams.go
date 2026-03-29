package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type TeamHandler struct {
	s *store.Store
}

func NewTeamHandler(s *store.Store) *TeamHandler {
	return &TeamHandler{s: s}
}

func (h *TeamHandler) Routes(r chi.Router) {
	r.Get("/teams", h.List)
	r.Get("/teams/new", h.NewForm)
	r.Post("/teams/new", h.Create)
	r.Get("/teams/{id}/edit", h.EditForm)
	r.Post("/teams/{id}/edit", h.Update)
	r.Post("/teams/{id}/delete", h.Delete)
	r.Get("/teams/export/csv", h.ExportCSV)
}

func (h *TeamHandler) APIRoutes(r chi.Router) {
	r.Get("/teams", h.APIList)
	r.Get("/teams/{id}", h.APIGet)
}

// HTML Handlers

func (th *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := th.s.ListTeams()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := view.TeamList(w, r, teams)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (th *TeamHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	page := view.TeamForm(nil, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (th *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data, errs := validation.ValidateTeam(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.TeamForm(nil, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err := th.s.PutTeam(&store.Team{
		Name:     data.Name,
		State:    data.State,
		IsActive: data.IsActive,
		Remarks:  data.Remarks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Team wurde erstellt.")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

func (th *TeamHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	team, err := th.getTeam(w, r)
	if err != nil {
		return
	}
	page := view.TeamForm(&team, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (th *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
	team, err := th.getTeam(w, r)
	if err != nil {
		return
	}

	r.ParseForm()
	data, errs := validation.ValidateTeam(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.TeamForm(&team, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	updated := &store.Team{
		ID:       team.ID,
		Name:     data.Name,
		State:    data.State,
		IsActive: data.IsActive,
		Remarks:  data.Remarks,
	}
	if err = th.s.PutTeam(updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update eingebettete Refs in allen Games
	th.s.UpdateGameRefs(func(g *store.Game) bool {
		changed := false
		if g.HomeTeam.ID == team.ID {
			g.HomeTeam = store.TeamRef{ID: updated.ID, Name: updated.Name}
			changed = true
		}
		if g.AwayTeam.ID == team.ID {
			g.AwayTeam = store.TeamRef{ID: updated.ID, Name: updated.Name}
			changed = true
		}
		return changed
	})

	view.SetFlash(w, "Team wurde aktualisiert.")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

func (th *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if th.s.IsReferenced(func(g *store.Game) bool { return g.HomeTeam.ID == id || g.AwayTeam.ID == id }) {
		view.SetFlash(w, "Team kann nicht gelöscht werden — wird noch in Spielen verwendet.")
		http.Redirect(w, r, "/teams", http.StatusSeeOther)
		return
	}

	if err := th.s.DeleteTeam(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Team wurde gelöscht.")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

// JSON API

func (th *TeamHandler) APIList(w http.ResponseWriter, r *http.Request) {
	teams, err := th.s.ListTeams()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, teams)
}

func (th *TeamHandler) APIGet(w http.ResponseWriter, r *http.Request) {
	team, err := th.getTeam(w, r)
	if err != nil {
		return
	}
	jsonResponse(w, team)
}

// Export

func (th *TeamHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	teams, err := th.s.ListTeams()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("teams", "csv"))
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Name", "Bundesland", "Aktiv", "Bemerkungen"})

	for _, t := range teams {
		active := "Nein"
		if t.IsActive {
			active = "Ja"
		}
		cw.Write([]string{t.Name, t.State, active, t.Remarks})
	}
	cw.Flush()
}

// Helpers

func (th *TeamHandler) getTeam(w http.ResponseWriter, r *http.Request) (store.Team, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return store.Team{}, fmt.Errorf("empty id")
	}
	team, err := th.s.GetTeam(id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "Team nicht gefunden", http.StatusNotFound)
		return store.Team{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return store.Team{}, err
	}
	return team, nil
}
