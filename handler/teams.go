package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type TeamHandler struct {
	q *model.Queries
}

func NewTeamHandler(q *model.Queries) *TeamHandler {
	return &TeamHandler{q: q}
}

func (h *TeamHandler) Routes(r chi.Router) {
	r.Get("/teams", h.List)
	r.Get("/teams/new", h.NewForm)
	r.Post("/teams/new", h.Create)
	r.Get("/teams/{id}/edit", h.EditForm)
	r.Post("/teams/{id}/edit", h.Update)
	r.Post("/teams/{id}/delete", h.Delete)
	r.Get("/teams/export/csv", h.ExportCSV)
	r.Get("/teams/export/sql", h.ExportSQL)
}

func (h *TeamHandler) APIRoutes(r chi.Router) {
	r.Get("/teams", h.APIList)
	r.Get("/teams/{id}", h.APIGet)
}

// HTML Handlers

func (th *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := th.q.ListTeams(r.Context())
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
	data, errors := validation.ValidateTeam(r.Form)

	if len(errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.TeamForm(nil, errors, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	_, err := th.q.CreateTeam(r.Context(), model.CreateTeamParams{
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
	data, errors := validation.ValidateTeam(r.Form)

	if len(errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.TeamForm(&team, errors, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err = th.q.UpdateTeam(r.Context(), model.UpdateTeamParams{
		ID:       team.ID,
		Name:     data.Name,
		State:    data.State,
		IsActive: data.IsActive,
		Remarks:  data.Remarks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Team wurde aktualisiert.")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

func (th *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := th.q.DeleteTeam(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Team wurde gelöscht.")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

// JSON API

func (th *TeamHandler) APIList(w http.ResponseWriter, r *http.Request) {
	teams, err := th.q.ListTeams(r.Context())
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
	teams, err := th.q.ListTeams(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="teams.csv"`)
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Name", "Bundesland", "Aktiv", "Bemerkungen"})

	for _, t := range teams {
		active := "Nein"
		if t.IsActive == 1 {
			active = "Ja"
		}
		cw.Write([]string{t.Name, t.State, active, t.Remarks})
	}
	cw.Flush()
}

func (th *TeamHandler) ExportSQL(w http.ResponseWriter, r *http.Request) {
	teams, err := th.q.ListTeams(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="teams.sql"`)

	for _, t := range teams {
		fmt.Fprintf(w, "INSERT INTO teams (name, state, is_active, remarks) VALUES (%s, %s, %d, %s);\n",
			sqlEscape(t.Name), sqlEscape(t.State), t.IsActive, sqlEscape(t.Remarks))
	}
}

// Helpers

func (th *TeamHandler) getTeam(w http.ResponseWriter, r *http.Request) (model.Team, error) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return model.Team{}, err
	}
	team, err := th.q.GetTeam(r.Context(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "Team nicht gefunden", http.StatusNotFound)
		return model.Team{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return model.Team{}, err
	}
	return team, nil
}
