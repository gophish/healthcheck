package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/healthcheck/db"
	"github.com/gophish/healthcheck/mail"
	"github.com/gophish/healthcheck/util"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewAPIRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.Timeout(60 * time.Second))

	// Setup CSRF Protection
	//r.Use(csrf.Protect([]byte(util.GenerateSecureID())))

	r.Route("/messages", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(RateLimit)
			r.Post("/", PostMessage)
		})
		r.Route("/{messageID}", func(r chi.Router) {
			r.Use(MessageCtx)
			r.Post("/{status}", UpdateMessage)
		})
	})

	return r
}

func MessageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the message from the database, updating the request context
		messageID := chi.URLParam(r, "messageID")
		message, err := db.GetMessage(messageID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "message", message)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimit is a function that limits requests to our POST endpoint by
// receiving domain. (TODO)
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func PostMessage(w http.ResponseWriter, r *http.Request) {
	m := db.Message{}
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.ErrorChan = make(chan error)
	hash, err := util.DomainHashFromAddress(m.Recipient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.DomainHash = hash
	db.PostMessage(&m)

	// Send the message to the mailer
	err = mail.SendEmail(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, m, http.StatusOK)
}

func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")
	w.Write([]byte(status))
}

// JSONResponse attempts to set the status code, c, and marshal the given interface, d, into a response that
// is written to the given ResponseWriter.
func JSONResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}
