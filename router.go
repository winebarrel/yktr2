package yktr2

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	goth_esa "github.com/winebarrel/goth-esa/esa"
	"github.com/winebarrel/yktr2/esa"
	"github.com/winebarrel/yktr2/utils"
)

type ContextKey string

const (
	SessionName               = "_yktr2_session"
	SessionUserKey            = "user"
	ContextUserKey ContextKey = "user"
)

//go:embed templates
var templates embed.FS

//go:embed favicon.ico
var favicon embed.FS

func init() {
	gothic.GetProviderName = func(req *http.Request) (string, error) {
		return "esa", nil
	}
}

func NewRouter(cfg *Config) *pat.Router {
	initGoth(cfg)
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	router := pat.New()
	router.Use(authorizeMiddleware(store))

	router.Get("/favicon.ico", http.FileServer(http.FS(favicon)).ServeHTTP)

	router.Get("/auth/callback", func(rw http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(rw, r)

		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(rw, err)
			return
		}

		sess, _ := store.Get(r, SessionName)
		sess.Values[SessionUserKey] = user
		sess.Save(r, rw)

		rw.Header().Set("Location", "/")
		rw.WriteHeader(http.StatusTemporaryRedirect)
	})

	router.Get("/logout", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)

		sess, _ := store.Get(req, SessionName)
		delete(sess.Values, SessionUserKey)
		sess.Save(req, res)
	})

	router.Get("/{path:(?:posts|members)}/{id}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		loc := fmt.Sprintf("http://%s.esa.io/%s/%s", cfg.Team, vars["path"], vars["id"])
		rw.Header().Set("Location", loc)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	})

	index := utils.NewTemplate(templates, "templates/index.html")
	postNotFound := utils.NewTemplate(templates, "templates/post_not_found.html")

	router.PathPrefix("/").Methods("GET").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(ContextUserKey).(goth.User)
		token := user.AccessToken

		q := r.URL.Query().Get("q")
		page := r.URL.Query().Get("page")
		category := strings.TrimLeft(r.URL.Path, "/")

		if category != "" {
			q = fmt.Sprintf(`%s in:"%s"`, q, category)
		}

		posts, err := esa.GetPosts(token, cfg.Team, q, page, strconv.Itoa(cfg.PerPage))

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(rw, err)
			return
		}

		data := map[string]interface{}{
			"q":          r.URL.Query().Get("q"),
			"category":   category,
			"domain":     fmt.Sprintf("%s.esa.io", cfg.Team),
			"stylesheet": esa.Stylesheet,
			"posts":      posts.Posts,
			"prev_page":  posts.PrevPage,
			"next_page":  posts.NextPage,
		}

		if len(posts.Posts) == 0 {
			postNotFound.Execute(rw, data)
			return
		}

		index.Execute(rw, data)
	})

	return router
}

func initGoth(cfg *Config) {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options.HttpOnly = true
	gothic.Store = store
	callback, _ := url.Parse(cfg.Oauth2.RedirectHost)
	callback.Path = path.Join(callback.Path, "auth/callback")

	goth.UseProviders(
		goth_esa.New(cfg.Oauth2.ClientID, cfg.Oauth2.ClientSecret, callback.String(), "read"),
	)
}

func authorizeMiddleware(store *sessions.CookieStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/auth") ||
				strings.HasPrefix(r.URL.Path, "/logout") ||
				strings.HasPrefix(r.URL.Path, "/favicon.ico") {
				next.ServeHTTP(rw, r)
				return
			}

			sess, _ := store.Get(r, SessionName)
			user := sess.Values[SessionUserKey]

			if user == nil {
				gothic.BeginAuthHandler(rw, r)
			} else {
				ctx := context.WithValue(r.Context(), ContextUserKey, user)
				next.ServeHTTP(rw, r.WithContext(ctx))
			}
		})
	}
}
