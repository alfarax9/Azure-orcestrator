package api

import (
	"net/http"

	"github.com/alfarax9/astra/internal/domain"
)

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"version": s.cfg.Version, "api": apiVersion})
}

// pair trades a pairing token for a device token. The device token appears in this answer and
// nowhere else; domain.Secret keeps it out of every other output, so it is revealed here.
func (s *Server) pair(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	d, token, err := s.auth.Pair(r.Context(), in.Token, in.Name)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"device": d, "token": token.Reveal()})
}

func (s *Server) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	list, err := s.reads.Workspaces(r.Context())
	respond(s, w, r, http.StatusOK, nonNil(list), err)
}

func (s *Server) addWorkspace(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	ws, err := s.ws.AddWorkspace(r.Context(), in.Slug, in.Name)
	respond(s, w, r, http.StatusCreated, ws, err)
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	ws, err := s.reads.Workspace(r.Context(), r.URL.Query().Get("workspace"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	list, err := s.reads.Projects(r.Context(), ws.ID)
	respond(s, w, r, http.StatusOK, nonNil(list), err)
}

func (s *Server) addProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Workspace string `json:"workspace"`
		Name      string `json:"name"`
		RepoURL   string `json:"repoUrl"`
	}
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	p, err := s.ws.AddProject(r.Context(), in.Workspace, in.Name, in.RepoURL)
	respond(s, w, r, http.StatusCreated, p, err)
}

func (s *Server) setCommands(w http.ResponseWriter, r *http.Request) {
	var in domain.Commands
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	id := domain.ProjectID(r.PathValue("id"))
	if err := s.ws.SetCommands(r.Context(), id, in); err != nil {
		s.fail(w, r, err)
		return
	}
	p, err := s.reads.Project(r.Context(), id)
	respond(s, w, r, http.StatusOK, p, err)
}

func (s *Server) listWorkers(w http.ResponseWriter, r *http.Request) {
	list, err := s.reads.Workers(r.Context(), r.URL.Query().Get("workspace"))
	respond(s, w, r, http.StatusOK, nonNil(list), err)
}

func (s *Server) getWorker(w http.ResponseWriter, r *http.Request) {
	wk, err := s.reads.Worker(r.Context(), domain.WorkerID(r.PathValue("id")))
	respond(s, w, r, http.StatusOK, wk, err)
}

func (s *Server) newWorker(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title        string           `json:"title"`
		Agent        domain.AgentKind `json:"agent"`
		ClientReview *bool            `json:"clientReview"`
	}
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	review := in.ClientReview == nil || *in.ClientReview // client review is the default (F6)
	wk, err := s.ws.CreateWorker(r.Context(), domain.ProjectID(r.PathValue("id")), in.Title, in.Agent, review)
	respond(s, w, r, http.StatusCreated, wk, err)
}

// startWorker runs the agent and keeps its facts true from then on.
func (s *Server) startWorker(w http.ResponseWriter, r *http.Request) {
	id := domain.WorkerID(r.PathValue("id"))
	if err := s.agents.Start(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	s.tracker.Track(id)
	wk, err := s.reads.Worker(r.Context(), id)
	respond(s, w, r, http.StatusOK, wk, err)
}

func (s *Server) stopWorker(w http.ResponseWriter, r *http.Request) {
	s.noContent(w, r, s.agents.Stop(r.Context(), domain.WorkerID(r.PathValue("id"))))
}

func (s *Server) interruptWorker(w http.ResponseWriter, r *http.Request) {
	s.noContent(w, r, s.agents.Interrupt(r.Context(), domain.WorkerID(r.PathValue("id"))))
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := decode(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	s.noContent(w, r, s.agents.Send(r.Context(), domain.WorkerID(r.PathValue("id")), in.Text))
}

func (s *Server) noContent(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// respond writes v, or the failure when err is set.
func respond(s *Server, w http.ResponseWriter, r *http.Request, status int, v any, err error) {
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, status, v)
}

// nonNil makes an empty list encode as [] rather than null.
func nonNil[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
}

func (s *Server) getChecklist(w http.ResponseWriter, r *http.Request) {
	c, err := s.list.Get(r.Context())
	respond(s, w, r, http.StatusOK, c, err)
}
