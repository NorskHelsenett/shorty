package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NorskHelsenett/shorty/internal/middleware"

	"github.com/NorskHelsenett/shorty/internal/models"
	redisdb "github.com/NorskHelsenett/shorty/internal/redis"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	AddAdminUser      = redisdb.AddAdminUser
	GetAllAdminEmails = redisdb.GetAllAdminEmails
	DeleteUser        = redisdb.DeleteUser
	AdminUserExists   = redisdb.AdminUserExists
)

// Add admin user
//
//		@Summary		Add admin user
//		@Schemes
//		@Description	adds a admin user
//		@Tags			v1 user
//		@Accept			application/json
//		@Produce		application/json
//		@Param			query	body		models.RedirectUser	true	"Query"
//		@Success		200		{object}	models.RedirectUser
//		@Failure		403		{string}	Forbidden
//		@Failure		401		{string}	Unauthorized
//	 	@Failure		409		{string}	Conflict
//		@Failure		500		{string}	Failure	message

// @Router			/v1/user [post]
// @Security		AccessToken
func AddUserRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		isAdmin := r.Context().Value(middleware.IsAdminKey).(bool)
		rlog.Debug("AddUserRedirect", rlog.Any("isAdmin", isAdmin))

		if !isAdmin {
			http.Error(w, "Forbidden: Only admin users can perform this action", http.StatusForbidden)
			rlog.Info("Forbidden: Only admin users can perform this action")
			// w.WriteHeader(http.StatusForbidden)
			return
		}

		defer r.Body.Close()

		resBody, err := io.ReadAll(r.Body)

		if err != nil {
			rlog.Error("impossible to read body of request", err)
			http.Error(w, "Impossible to read body of request", http.StatusInternalServerError)
			return
		}

		var res models.RedirectUser // object res (userID, lastname ect..)

		// Deserialize JSON to the Go model 'Redirect'
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			rlog.Error("impossible to unmarshal body of request", err)
			http.Error(w, "Impossible to unmarshal body of request", http.StatusNotAcceptable)
			return
		}

		// Checks the email is empty
		if res.Email == "" {
			rlog.Info("email is empty")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		verifier := emailverifier.NewVerifier()

		ver, err := verifier.Verify(res.Email)
		if err != nil || !ver.Syntax.Valid {
			rlog.Info("Invalid email format")
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}

		userID := uuid.New().String()

		status, err := AddAdminUser(rdb, userID, res.Email)
		if err != nil {
			rlog.Error("Error: AddUser failed", err)
			http.Error(w, "Error occurred while adding/updating user", http.StatusBadRequest)
			return
		}

		if status == "exists" {
			rlog.Info("Admin user already exists")
			ret := models.Response{
				Success: false,
				Message: "Admin user already exists, email: " + res.Email,
			}
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(ret); err != nil {
				rlog.Error("Error encoding response: ", err)
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
			return
		}

		responsMessage := fmt.Sprintf("User %s successfuly: %s", status, res.Email)
		rlog.Info(responsMessage, rlog.Any("status", status), rlog.Any("email", res.Email))

		ret :=
			models.Response{
				Success: true,
				Message: responsMessage,
			}
		ret.Message = fmt.Sprintf(`User %s : %s`, status, res.Email)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			rlog.Error("Error encoding response: ", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Get all admin users
//
//	@Summary		Get all admin
//	@Schemes
//	@Description	Returns a list of all redirect entries configured for the admin panel.
//	@Tags			admin user
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	[]models.RedirectUser
//	@Failure		403	{string}	Forbidden
//	@Failure		401	{string}	Unauthorized
//	@Failure		500	{string}	Failure	message
//	@Router			/v1/user [get]
//	@Security		AccessToken
func GetAllUsersRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		isAdmin := r.Context().Value(middleware.IsAdminKey).(bool)
		rlog.Info("GetAllUserRedirect", rlog.Any("isAdmin: ", isAdmin))

		if !isAdmin {
			http.Error(w, "Forbidden: Only admin users can perform this action", http.StatusForbidden)
			rlog.Info("Forbidden: Only admin users can perform this action")
			// w.WriteHeader(http.StatusForbidden)
			return
		}

		defer r.Body.Close()

		redirects, err := GetAllAdminEmails(rdb)

		if err != nil {
			rlog.Error("Error reading admin emails", err)
			http.Error(w, "Error: Unable to retrieve admin emails", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(redirects); err != nil {
			rlog.Error("Error encoding response: ", err)
			http.Error(w, "Error: Unable to encode response", http.StatusInternalServerError)
			return
		}

	}
}

// Delete admin user
//
//	@Summary	Delete admin user
//	@Schemes
//	@Description	deletes a admin user by email
//	@Tags			v1 user
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string	true	"Id"
//	@Success		200	{object}	models.RedirectUser
//	@Failure		403	{string}	Forbidden
//	@Failure		401	{string}	Unauthorized
//	@Failure		500	{string}	Failure	message
//	@Router			/v1/user/{id} [delete]
//	@Security		AccessToken
func DeleteUserRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rlog.Info("DeleteUserRedirect: Handler called")

		isAdmin, adminOk := r.Context().Value(middleware.IsAdminKey).(bool)
		if !adminOk {
			rlog.Warn("DeleteUserRedirect: Missing admin status in context")
			http.Error(w, "Unauthorized: Admin status is missing in context", http.StatusUnauthorized)
			return
		}
		if !isAdmin {
			rlog.Warn("DeleteUserRedirect: User is not an admin, access denied")
			http.Error(w, "Forbidden: Only admin users can perform this action", http.StatusForbidden)
			return
		}

		email := mux.Vars(r)["id"]
		if email == "" {
			rlog.Warn("DeleteUserRedirect: Missing email parameter in URL")
			http.Error(w, "Email is required in URL", http.StatusBadRequest)
			return
		}

		err := DeleteUser(rdb, email)
		if err != nil {
			rlog.Error("DeleteUserRedirect: Failed to delete user", err)
			if err.Error() == "email not found" {
				http.Error(w, "Email not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal Server Error: Failed to delete user", http.StatusInternalServerError)
			}
			return
		}

		rlog.Info("DeleteUserRedirect: User deleted successfully", rlog.String("email", email))
		w.WriteHeader(http.StatusOK)
		jsonResponse := map[string]string{"message": "User deleted successfully"}
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			rlog.Error("Error encoding response: ", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func CheckUserEmailRedirect(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAdmin := r.Context().Value(middleware.IsAdminKey).(bool)
		rlog.Debug("AddUserRedirect", rlog.Any("isAdmin", isAdmin))

		if !isAdmin {
			http.Error(w, "Forbidden: Only admin users can perform this action", http.StatusForbidden)
			rlog.Info("Forbidden: Only admin users can perform this action")
			// w.WriteHeader(http.StatusForbidden)
			return
		}

		defer r.Body.Close()

		resBody, err := io.ReadAll(r.Body)

		if err != nil {
			rlog.Error("impossible to read body of request", err)
			http.Error(w, "Impossible to read body of request", http.StatusInternalServerError)
			return
		}

		var res models.RedirectUser // object res (userID, email..)

		// Deserialize JSON to the Go model 'Redirect'
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			rlog.Error("impossible to unmarshal body of request", err)
			http.Error(w, "Impossible to unmarshal body of request", http.StatusNotAcceptable)
			return
		}

		// Checks the email is empty
		if res.Email == "" {
			rlog.Info("email is empty")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		emailExists := AdminUserExists(rdb, res.Email)

		response := map[string]interface{}{
			"Exists": emailExists,
			"email":  res.Email,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			rlog.Error("Error encoding response: ", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
