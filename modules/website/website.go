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
	Host       string
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
	loginHandler := newLoginHandler(w.config, "/login")

	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandlerFunc)
	mux.HandleFunc(loginHandler.path, loginHandler.handle)

	err := http.ListenAndServe(fmt.Sprintf(":%d", w.config.Port), mux)
	if err != nil {
		return err
	}

	return nil
}

func HomeHandlerFunc(w http.ResponseWriter, _ *http.Request) {
	homeTemplate.Execute(w, nil)
}

type loginHandler struct {
	config      WebsiteConfig
	path        string
	authHandler *auth.AuthHandler
}

func newLoginHandler(config WebsiteConfig, path string) *loginHandler {
	return &loginHandler{
		config:      config,
		path:        path,
		authHandler: auth.NewAuthHandler(path, config.Host, config.Port, config.AuthConfig.SteamKey),
	}
}

func (h *loginHandler) handle(w http.ResponseWriter, r *http.Request) {
	profile, err := h.authHandler.Handle(w, r)
	if _, redirect := err.(*auth.Redirect); redirect {
		return
	}

	type loginModel struct {
		PersonaName string
		Error       string
	}

	var model loginModel

	if profile != nil {
		model.PersonaName = profile.PersonaName
	}

	if err != nil {
		model.Error = err.Error()
	}

	loginTemplate.Execute(w, model)
}
