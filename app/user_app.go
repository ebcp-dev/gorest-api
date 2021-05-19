package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	app "github.com/ebcp-dev/gorest-api/app/utils"
	"github.com/ebcp-dev/gorest-api/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Used for validating header tokens.
var mySigningKey = []byte("captainjacksparrowsayshi")

// Initialize DB and routes.
func (a *App) UserInitialize() {
	a.initializeUserRoutes()
}

// Defines routes.
func (a *App) initializeUserRoutes() {
	a.Router.HandleFunc("/user", a.createUser).Methods("POST")
	a.Router.HandleFunc("/user/login", a.loginUser).Methods("POST")
	// Authorized routes.
	a.Router.Handle("/user/{id}", a.isAuthorized(a.getUser)).Methods("GET")
	a.Router.Handle("/users", a.isAuthorized(a.getUsers)).Methods("GET")
	a.Router.Handle("/user/{id}", a.isAuthorized(a.updateUser)).Methods("PUT")
	a.Router.Handle("/user/{id}", a.isAuthorized(a.deleteUser)).Methods("DELETE")
}

// Route handlers

// Retrieves user from db using id from URL.
func (a *App) loginUser(w http.ResponseWriter, r *http.Request) {
	var u model.User
	// Gets JSON object from request body.
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		app.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()
	// Find user in db with email and password from request body.
	if err := u.GetUserByEmailAndPassword(d.Database); err != nil {
		switch err {
		case sql.ErrNoRows:
			// Respond with 404 if user not found in db.
			app.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			// Respond if internal server error.
			app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// Generate and send token to client with response header.
	validToken, err := GenerateJWT()
	if err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Add("Token", validToken)
	// Respond with user in db.
	app.RespondWithJSON(w, http.StatusOK, u)
}

// Retrieves user from db using id from URL.
func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Convert id string variable to int.
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	u := model.User{ID: id}
	if err := u.GetUser(d.Database); err != nil {
		switch err {
		case sql.ErrNoRows:
			// Respond with 404 if user not found in db.
			app.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			// Respond if internal server error.
			app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// If user found respond with user object.
	app.RespondWithJSON(w, http.StatusOK, u)
}

// Gets list of user with count and start variables from URL.
func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	// Convert count and start string variables to int.
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	// Default and limit of count is 10.
	if count > 10 || count < 1 {
		count = 10
	}
	// Min start is 0;
	if start < 0 {
		start = 0
	}

	users, err := model.GetUsers(d.Database, start, count)
	if err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	app.RespondWithJSON(w, http.StatusOK, users)
}

// Inserts new user into db.
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u model.User
	// Gets JSON object from request body.
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		app.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	if err := u.CreateUser(d.Database); err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Respond with newly created user.
	app.RespondWithJSON(w, http.StatusCreated, u)
}

// Updates user in db using id from URL.
func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Convert id string variable to int.
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	var u model.User
	// Gets JSON object from request body.
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		app.RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}

	defer r.Body.Close()
	u.ID = id

	if err := u.UpdateUser(d.Database); err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Respond with updated user.
	app.RespondWithJSON(w, http.StatusOK, u)
}

// Deletes user in db using id from URL.
func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Convert id string variable to int.
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	u := model.User{ID: id}
	if err := u.DeleteUser(d.Database); err != nil {
		app.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Respond with success message if operation is completed.
	app.RespondWithJSON(w, http.StatusOK, map[string]string{"result": "User deleted"})
}

// Helper functions

// Generate JWT
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = "Elliot Forbes"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		// fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}
