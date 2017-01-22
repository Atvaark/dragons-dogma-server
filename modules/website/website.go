package website

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/atvaark/dragons-dogma-server/modules/auth"
)

type Website struct {
	config WebsiteConfig
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

func NewWebsite(cfg WebsiteConfig) *Website {
	return &Website{cfg}
}

func (w *Website) ListenAndServe() error {
	sessionStore := auth.NewSessionStore()
	sessionHandler := auth.NewSessionHandler(sessionStore)
	authHandler := auth.NewAuthHandler(w.config.RootURL, "/login/", w.config.AuthConfig.SteamKey)
	homeHandler := &homeHandler{w.config.RootURL, "/", sessionHandler}
	loginHandler := &loginHandler{w.config.RootURL, "/login/", sessionHandler, authHandler}

	mux := http.NewServeMux()
	mux.HandleFunc(homeHandler.path, homeHandler.handle)
	mux.HandleFunc(loginHandler.path, loginHandler.handle)

	err := http.ListenAndServe(fmt.Sprintf(":%d", w.config.Port), mux)
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
	profile, _ := h.sessionHandler.GetSessionCookie(w, r)

	var model homeModel
	model.RootURL = h.rootURL

	if profile != nil {
		model.PersonaName = profile.PersonaName
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
	profile, loggedIn := h.sessionHandler.GetSessionCookie(w, r)
	if !loggedIn {
		profile, err = h.authHandler.Handle(w, r)
		if _, redirect := err.(*auth.Redirect); redirect {
			return
		}

		if err == nil {
			h.sessionHandler.SetSessionCookie(w, profile)
		}
	}

	var model loginModel
	model.RootURL = h.rootURL

	if profile != nil {
		model.PersonaName = profile.PersonaName
	}

	if err != nil {
		model.Error = err.Error()
	}

	loginTemplate.Execute(w, model)
}
