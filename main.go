package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// RequestPayload represents the expected request structure
type RequestPayload struct {
	FlowID     string `json:"flowId"`
	RequestID  string `json:"requestId"`
	ActionType string `json:"actionType"`
	Event      struct {
		InitiatorType string `json:"initiatorType"`
		Action        string `json:"action"`
		Tenant        struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tenant"`
		User struct {
			ID                 string `json:"id"`
			UpdatingCredential struct {
				Type           string `json:"type"`
				Format         string `json:"format"`
				Value          string `json:"value"`
				AdditionalData struct {
					Algorithm string `json:"algorithm"`
				} `json:"additionalData"`
			} `json:"updatingCredential"`
		} `json:"user"`
		UserStore struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"userStore"`
	} `json:"event"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	ActionStatus string `json:"actionStatus"`
}

// FailedResponse represents a failed response but with HTTP 200
type FailedResponse struct {
	ActionStatus       string `json:"actionStatus"`
	FailureReason      string `json:"failureReason"`
	FailureDescription string `json:"failureDescription"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	ActionStatus     string `json:"actionStatus"`
	Error            string `json:"error"`
	ErrorDescription string `json:"errorDescription"`
}

// Load common passwords from a file
func loadCommonPasswords(filename string) (map[string]bool, error) {
	commonPasswords := make(map[string]bool)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		commonPasswords[strings.TrimSpace(scanner.Text())] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commonPasswords, nil
}

// sendJSONResponse is a helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"actionStatus":"ERROR","error":"encoding_failed","errorDescription":"Failed to encode JSON"}`, http.StatusInternalServerError)
	}
}

func passwordValidationHandler(commonPasswords map[string]bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the response is always JSON
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			sendJSONResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
				ActionStatus:     "ERROR",
				Error:            "invalid_request",
				ErrorDescription: "Invalid request method",
			})
			return
		}

		var payload RequestPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			sendJSONResponse(w, http.StatusBadRequest, ErrorResponse{
				ActionStatus:     "ERROR",
				Error:            "invalid_payload",
				ErrorDescription: "Invalid request payload",
			})
			return
		}

		password := payload.Event.User.UpdatingCredential.Value

		// Check if password is in the common passwords list
		if commonPasswords[password] {
			sendJSONResponse(w, http.StatusOK, FailedResponse{
				ActionStatus:       "FAILED",
				FailureReason:      "common_password",
				FailureDescription: "The provided password is a common password.",
			})
			return
		}

		// Success response if password is valid
		sendJSONResponse(w, http.StatusOK, SuccessResponse{
			ActionStatus: "SUCCESS",
		})
	}
}

func main() {
	// Load common passwords from "resources/common-passwords.txt"
	commonPasswords, err := loadCommonPasswords("resources/common-passwords.txt")
	if err != nil {
		fmt.Println("Error loading common passwords:", err)
		return
	}

	http.HandleFunc("/", passwordValidationHandler(commonPasswords))

	port := 8080
	fmt.Printf("Server is running on port %d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
