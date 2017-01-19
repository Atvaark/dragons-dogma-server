package website

import (
	"fmt"
	"html/template"
	"net/http"
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
	loginHandler := &loginHandler{config: w.config, path: "/login/"}

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

type loginModel struct {
	PersonaName string
}

type loginHandler struct {
	config WebsiteConfig
	path   string
}

func (h *loginHandler) handle(w http.ResponseWriter, r *http.Request) {
	openid, openidFound := parseOpenid(r.URL.Query())
	if !openidFound {
		// TODO: get protocol/port from config
		callbackURL := fmt.Sprintf("http://%s:%d%s", h.config.Host, h.config.Port, h.path)

		steamLogin, err := buildAuthUrl(callbackURL)
		if err != nil {
			http.Error(w, "could not initialize steam login", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, steamLogin, http.StatusTemporaryRedirect)
		return
	}

	steamId, err := validateOpenid(openid)
	if err != nil {
		http.Error(w, "could not validate steam login", http.StatusInternalServerError)
		return
	}

	profile, err := fetchUserProfile(h.config.AuthConfig.SteamKey, steamId)
	if err != nil {
		http.Error(w, "could not fetch user profile", http.StatusInternalServerError)
		return
	}

	// TODO: create login session

	model := loginModel{
		PersonaName: profile.PersonaName,
	}

	loginTemplate.Execute(w, model)
}
