package website

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/atvaark/dragons-dogma-server/modules/auth"
)

type Website struct {
	server *http.Server
}

type AuthConfig struct {
	SteamKey string
}

type WebsiteConfig struct {
	RootURL    string
	Port       int
	AuthConfig AuthConfig
}

var (
	homeTemplate  = template.Must(template.New("home.tmpl").ParseFiles("templates/home.tmpl"))
	loginTemplate = template.Must(template.New("login.tmpl").ParseFiles("templates/login.tmpl"))
)

func NewWebsite(cfg WebsiteConfig, database auth.Database) *Website {
	sessionHandler := auth.NewSessionHandler(database)
	authHandler := auth.NewAuthHandler(cfg.RootURL, "/login/", cfg.AuthConfig.SteamKey)
	homeHandler := &homeHandler{cfg.RootURL, "/", sessionHandler}
	loginHandler := &loginHandler{cfg.RootURL, "/login/", sessionHandler, authHandler}

	mux := http.NewServeMux()
	mux.HandleFunc(homeHandler.path, homeHandler.handle)
	mux.HandleFunc(loginHandler.path, loginHandler.handle)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return &Website{
		server: srv,
	}
}

func (w *Website) ListenAndServe() error {
	err := w.server.ListenAndServe()
	if err != nil && err.Error() != "http: Server closed" {
		return err
	}

	return nil
}

func (w *Website) Close() error {
	var ctx context.Context
	err := w.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

type rootModel struct {
	RootURL string
}

type homeHandler struct {
	rootURL        string
	path           string
	sessionHandler *auth.SessionHandler
}

type homeModel struct {
	rootModel
	PersonaName string
	LoggedIn    bool
}

func (h *homeHandler) handle(w http.ResponseWriter, r *http.Request) {
	user, _ := h.sessionHandler.GetSessionCookie(w, r)

	var model homeModel
	model.RootURL = h.rootURL

	if user != nil {
		model.PersonaName = user.PersonaName
		model.LoggedIn = true
	}

	homeTemplate.Execute(w, model)
}

type loginHandler struct {
	rootURL        string
	path           string
	sessionHandler *auth.SessionHandler
	authHandler    *auth.AuthHandler
}

type loginModel struct {
	rootModel
	PersonaName string
	Error       string
}

func (h *loginHandler) handle(w http.ResponseWriter, r *http.Request) {
	var err error
	user, loggedIn := h.sessionHandler.GetSessionCookie(w, r)
	if !loggedIn {
		user, err = h.authHandler.Handle(w, r)
		if _, redirect := err.(*auth.Redirect); redirect {
			return
		}

		if err == nil {
			h.sessionHandler.SetSessionCookie(w, user)
		}
	}

	var model loginModel
	model.RootURL = h.rootURL

	if user != nil {
		model.PersonaName = user.PersonaName
	}

	if err != nil {
		model.Error = err.Error()
	}

	loginTemplate.Execute(w, model)
}
