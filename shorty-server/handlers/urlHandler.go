package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"shorty/metrics"
	"shorty/middleware"
	"shorty/models"
	redisdb "shorty/redis"
	"strings"
	"time"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// CheckURL validates if a URL with the given ID exists
// Returns (exists, statusCode, errorMessage)
func CheckURL(rdb *redis.Client, id string) (bool, int, string) {
	if id == "" {
		return false, http.StatusBadRequest, "Missing key parameter"
	}

	exists, err := redisdb.URLExists(rdb, id)
	if err != nil {
		rlog.Error("Error checking if URL exists", err, rlog.Any("id", id))
		return false, http.StatusInternalServerError, "Internal server error"
	}

	if !exists {
		return false, http.StatusNotFound, "URL does not exist"
	}
	return true, http.StatusOK, ""
}

// Redirect to URL
//
//	@Summary	Redirect
//	@Schemes
//	@Description	redirects to the URL
//	@Tags			redirect
//	@Accept			text/html
//	@Produce		text/html
//	@Param			path	path		string	true	"Path"
//	@Success		302		{string}	Redirecting
//	@Failure		403		{string}	Forbidden
//	@Failure		401		{string}	Unauthorized
//	@Failure		500		{string}	Failure	message
//	@Router			/{path} [get]
func Redirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		path, err := redisdb.GetURL(rdb, id)

		rlog.Info("Redirect", rlog.Any("id", id))

		if err != nil {
			path = "https://nhn.no"
			rlog.Info("Default redirect, path not found", rlog.Any("client", r.Host), rlog.Any("path", r.RequestURI), rlog.Any("to", path))
			http.Redirect(w, r, path, http.StatusFound)
			return
		}
		rlog.Info("Redirecting", rlog.Any("client", r.Host), rlog.Any("path", r.RequestURI), rlog.Any("to", path))

		// increment httpRequest metric on path
		currentYearMonth := time.Now().Format("2006-01")
		metrics.RequestCount.WithLabelValues(r.URL.Path, currentYearMonth).Inc()

		http.Redirect(w, r, path, http.StatusFound)
	}
}

// Delete redirect
//
//	@Summary	Delete redirect
//	@Schemes
//	@Description	delets a redirect by id
//	@Tags			admin
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string	true	"Id"
//	@Success		200	{object}	models.Response
//	@Failure		403	{string}	Forbidden
//	@Failure		401	{string}	Unauthorized
//	@Failure		500	{string}	Failure	message
//	@Router			/admin/{id} [delete]
//	@Security		AccessToken
func DeleteRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("deleteRedirect called")

		isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
		isOwner, _ := r.Context().Value(middleware.IsOwnerKey).(bool)
		rlog.Debug("isOwnerOrAdmin", rlog.Any("isAdminUser:", isAdmin))
		rlog.Debug("isOwnerOrAdmin", rlog.Any("isOwner", isOwner))

		if !isAdmin && !isOwner {
			w.WriteHeader(http.StatusForbidden)
			http.Error(w, "Forbidden: You must be an admin or the owner of this resource", http.StatusForbidden)
			return
		}

		params := mux.Vars(r) // get variable from request
		id := params["id"]

		if ok, statusCode, msg := CheckURL(rdb, id); !ok {
			http.Error(w, msg, statusCode)
			return
		}

		success, err := redisdb.Delete(rdb, id)
		if !success || err != nil {
			rlog.Error("Failed to delete URl", err)
			http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
			return
		}

		rlog.Info("URL deleted successfully")
		w.WriteHeader(http.StatusOK)
		jsonResponse := map[string]string{"message": "Path deleted successfully"}
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			rlog.Error("Error encoding response: ", err)
			return
		}
	}
}

// Update redirect
//
//	@Summary	Updates redirect
//	@Schemes
//	@Description	Updates a redirect to given url
//	@Tags			admin
//	@Accept			application/json
//	@Produce		application/json
//	@Param			query	body		models.Redirect	true	"Query"
//	@Param			id		path		string			true	"Id"
//	@Success		200		{object}	models.Response
//	@Failure		403		{string}	Forbidden
//	@Failure		401		{string}	Unauthorized
//	@Failure		500		{string}	Failure	message
//	@Router			/admin/{id} [patch]
//	@Security		AccessToken
func UpdateRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rlog.Debug("UpdateRedirect called")

		isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
		isOwner, _ := r.Context().Value(middleware.IsOwnerKey).(bool)
		rlog.Debug("isOwnerOrAdmin", rlog.Any("isAdminUser:", isAdmin))
		rlog.Debug("isOwnerOrAdmin", rlog.Any("isOwner", isOwner))

		if !isAdmin && !isOwner {
			w.WriteHeader(http.StatusForbidden)
			http.Error(w, "Forbidden: You must be an admin or the owner of this resource", http.StatusForbidden)
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if ok, _, msg := CheckURL(rdb, id); !ok {
			rlog.Info(msg)
			return
		}

		lastEditedBy, ok := r.Context().Value(middleware.UserKey).(string)
		if !ok || lastEditedBy == "" {
			rlog.Warn("Failed to retrieve user userEmail form context")
			http.Error(w, "user userEmail not found", http.StatusUnauthorized)
			return
		}

		// decode body
		var update models.Redirect
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			rlog.Error("Failed to decode body", err)
			http.Error(w, "Failed to decode body", http.StatusBadRequest)
			return
		}

		// update URL in Redis
		_, err := redisdb.UpdateOrCreatePath(rdb, id, update.URL, lastEditedBy)
		if err != nil {
			http.Error(w, "Failed to update URL", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		jsonResponse := map[string]string{"message": "Path deleted successfully"}
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			rlog.Error("Error encoding response: ", err)
			return
		}
	}
}

// Add redirect
//
//	@Summary	Add redirect
//	@Schemes
//	@Description	adds a redirect to url
//	@Tags			admin
//	@Accept			application/json
//	@Produce		application/json
//	@Param			query	body		models.Redirect	true	"Query"
//	@Success		200		{object}	models.Response
//	@Failure		403	{string}	Forbidden
//	@Failure		401	{string}	Unauthorized
//	@Failure		500	{string}	Failure	message
//	@Router			/admin/ [post]
//	@Security		AccessToken
func AddRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rlog.Debug("AddRedirect called")

		var redirect models.Redirect
		if err := json.NewDecoder(r.Body).Decode(&redirect); err != nil {
			rlog.Error("Failed to decode request body", err)
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Validate URL
		if redirect.URL == "" {
			rlog.Info("URL is empty")
			http.Error(w, "URL cannot be empty", http.StatusBadRequest)
			return
		}

		if !IsURL(redirect.URL) {
			rlog.Info("Invalid URL format")
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		// Check if path already exists
		exists, err := redisdb.URLExists(rdb, redirect.Path)
		if err != nil {
			rlog.Error("Failed to check if URL exists", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if exists {
			rlog.Info("Path already exists", rlog.Any("path", redirect.Path))
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(models.Response{
				Success: false,
				Message: fmt.Sprintf("Path already exists: %s", redirect.Path),
			})

			if err != nil {
				rlog.Error("Error encoding response: ", err)
				return
			}

			return
		}

		// Get user from context
		userEmail, ok := r.Context().Value(middleware.UserKey).(string)
		if !ok || userEmail == "" {
			rlog.Warn("User email not found in context")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Create the redirect
		message, err := redisdb.UpdateOrCreatePath(rdb, redirect.Path, redirect.URL, userEmail)
		if err != nil {
			rlog.Error("Failed to create redirect", err)
			http.Error(w, "Failed to create redirect", http.StatusInternalServerError)
			return
		}

		// Send successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(models.Response{
			Success: true,
			Message: message,
		}); err != nil {
			rlog.Error("Failed to encode response", err)
		}
	}
}

// Get all redirects
//
//	@Summary	Get redirect
//	@Schemes
//	@Description	gets all redirects
//	@Tags			admin
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	[]models.Redirect
//	@Failure		403	{string}	Forbidden
//	@Failure		401	{string}	Unauthorized
//	@Failure		500	{string}	Failure	message
//	@Router			/admin/ [get]
//	@Security		AccessToken
func GetAllRedirects(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, _ := r.Context().Value(middleware.UserKey).(string)
		isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)

		redirects, err := redisdb.GetAll(rdb, "path")
		if err != nil {
			rlog.Error("Error in GetAll: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(redirects) == 0 {
			redirects = []models.RedirectPath{}
		}

		var redirectsMap []models.RedirectAllPaths

		for _, redirect := range redirects {

			isOwner := redirect.Owner == user
			canModify := isOwner || isAdmin

			redirectsMap = append(redirectsMap, models.RedirectAllPaths{
				Path:   redirect.Path,
				URL:    redirect.URL,
				Owner:  redirect.Owner,
				Modify: canModify,
			})
		}

		w.Header().Set("Content-Type", "application/json")

		jssonstring, err := json.Marshal(redirectsMap)
		if err != nil {
			rlog.Error("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(jssonstring); err != nil {
			rlog.Error("Failed to write response: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

// IsURL validates if a string is a properly formatted URL
// Returns true if the string is a valid URL, false otherwise
func IsURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		rlog.Info("Invalid URL format", rlog.Any("error", err.Error()))
		return false
	}

	// Check if host is an IP address
	address := net.ParseIP(u.Host)
	if address == nil {
		return strings.Contains(u.Host, ".")
	}

	return true
}
