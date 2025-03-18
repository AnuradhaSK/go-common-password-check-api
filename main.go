package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// FailedResponse represents a failed response
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

// sendJSONResponse is a helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"actionStatus":"ERROR","error":"encoding_failed","errorDescription":"Failed to encode JSON"}`, http.StatusInternalServerError)
	}
}

func passwordValidationHandler(w http.ResponseWriter, r *http.Request) {
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

	// Check if password is one of the disallowed values
	if password == "admin123" || password == "password123" {
		sendJSONResponse(w, http.StatusOK, FailedResponse{
			ActionStatus:       "FAILED",
			FailureReason:      "password_compromised",
			FailureDescription: "The provided password is compromised.",
		})
		return
	}

	// Success response if password is valid
	sendJSONResponse(w, http.StatusOK, SuccessResponse{
		ActionStatus: "SUCCESS",
	})
}

func main() {
	http.HandleFunc("/", passwordValidationHandler)

	port := 8080
	fmt.Printf("Server is running on port %d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
