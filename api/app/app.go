package app

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	app "github.com/ebcp-dev/gorest-api/api/app/utils"
	"github.com/ebcp-dev/gorest-api/api/db"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// References DB struct in db.go.
var d db.DB

type App struct {
	Router *mux.Router
}

// Initialize DB and routes.
func (a *App) Initialize() {
	// Get env variables.
	db_user := os.Getenv("APP_DB_USERNAME")
	db_pass := os.Getenv("APP_DB_PASSWORD")
	db_host := os.Getenv("APP_DB_HOST")
	db_name := os.Getenv("APP_DB_NAME")
	if os.Getenv("ENV") == "prod" {
		db_user = os.Getenv("PROD_DB_USERNAME")
		db_pass = os.Getenv("PROD_DB_PASSWORD")
		db_host = os.Getenv("PROD_DB_HOST")
		db_name = os.Getenv("PROD_DB_NAME")
	}
	// Receives database credentials and connects to database.
	d.Initialize(db_user, db_pass, db_host, db_name)

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/", homePage)
	a.UserInitialize()
	a.DataInitialize()
}

// Serve homepage
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to GoREST - API")
}

// Starts the application.
func (a *App) Run(addr string) {
	log.Printf("Server listening on port: %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// Authorization middleware
func (a *App) isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request has "Token" header.
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				// Check if token is valid based on private `mySigningKey`.
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					app.RespondWithError(w, http.StatusInternalServerError, "There was error with signing the token.")
				}
				return mySigningKey, nil
			})

			if err != nil {
				app.RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
			// Serve endpoint if token is valid.
			if token.Valid {
				endpoint(w, r)
			}
		} else {
			app.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		}
	})
}