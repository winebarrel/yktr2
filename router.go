package yktr2

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	goth_esa "github.com/winebarrel/goth-esa/esa"
	"github.com/winebarrel/yktr2/esa"
	"github.com/winebarrel/yktr2/utils"
)

type ContextKey string

const (
	sessionName               = "_yktr2_session"
	sessionUserKey            = "user"
	contextUserKey ContextKey = "user"
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

func NewRouter(cfg *Config) http.Handler {
	initGoth(cfg)
	store := newCookieStore(cfg.SessionSecret, cfg.CookieSecure)
	router := mux.NewRouter()
	router.Use(authorizeMiddleware(store))

	router.Path("/favicon.ico").Methods("GET").Handler(http.FileServer(http.FS(favicon)))

	router.Path("/ping").Methods("GET").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, "PONG")
	})

	router.Path("/auth/esa/callback").Methods("GET").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(rw, r)

		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(rw, err)
			return
		}

		sess, _ := store.Get(r, sessionName)
		sess.Values[sessionUserKey] = user
		sess.Save(r, rw)

		rw.Header().Set("Location", "/")
		rw.WriteHeader(http.StatusTemporaryRedirect)
	})

	router.Path("/logout").Methods("GET").HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)

		sess, _ := store.Get(req, sessionName)
		delete(sess.Values, sessionUserKey)
		sess.Save(req, res)
	})

	router.Path("/{path:(?:posts|members)}/{id:.+}").Methods("GET").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		loc := fmt.Sprintf("http://%s.esa.io/%s/%s", cfg.Team, vars["path"], vars["id"])
		rw.Header().Set("Location", loc)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	})

	index := utils.NewTemplate(templates, "templates/index.html")
	postNotFound := utils.NewTemplate(templates, "templates/post_not_found.html")
	esaCli := esa.NewClient(cfg.Team, cfg.PerPage)

	router.PathPrefix("/").Methods("GET").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		token := user.AccessToken

		q := r.URL.Query().Get("q")
		page := r.URL.Query().Get("page")
		category := strings.TrimLeft(r.URL.Path, "/")

		if category != "" {
			q = fmt.Sprintf(`%s in:"%s"`, q, category)
		}

		posts, err := esaCli.Posts(token, q, page)

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

	return handlers.LoggingHandler(os.Stdout, router)
}

func initGoth(cfg *Config) {
	gothic.Store = newCookieStore(cfg.SessionSecret, cfg.CookieSecure)
	callback, _ := url.Parse(cfg.Oauth2.RedirectHost)
	callback.Path = path.Join(callback.Path, "auth/esa/callback")

	goth.UseProviders(
		goth_esa.New(cfg.Oauth2.ClientID, cfg.Oauth2.ClientSecret, callback.String(), "read"),
	)
}

func authorizeMiddleware(store *sessions.CookieStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/auth") ||
				strings.HasPrefix(r.URL.Path, "/logout") ||
				strings.HasPrefix(r.URL.Path, "/favicon.ico") ||
				strings.HasPrefix(r.URL.Path, "/ping") {
				next.ServeHTTP(rw, r)
				return
			}

			sess, _ := store.Get(r, sessionName)
			user := sess.Values[sessionUserKey]

			if user == nil {
				gothic.BeginAuthHandler(rw, r)
			} else {
				ctx := context.WithValue(r.Context(), contextUserKey, user)
				next.ServeHTTP(rw, r.WithContext(ctx))
			}
		})
	}
}

func newCookieStore(secret string, secure bool) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options.HttpOnly = true
	store.Options.MaxAge = 86400 * 30
	store.Options.Secure = secure
	return store
}

func getUser(r *http.Request) goth.User {
	return r.Context().Value(contextUserKey).(goth.User)
}
