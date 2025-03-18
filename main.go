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
	ActionStatus    string `json:"actionStatus"`
	ErrorMessage    string `json:"error"`
	ErrorDescription string `json:"errorDescription"`
}

func passwordValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			ActionStatus:    "ERROR",
			ErrorMessage:    "invalid_request",
			ErrorDescription: "Invalid request method",
		})
		return
	}

	var payload RequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			ActionStatus:    "ERROR",
			ErrorMessage:    "invalid_payload",
			ErrorDescription: "Invalid request payload",
		})
		return
	}

	password := payload.Event.User.UpdatingCredential.Value

	// Check if password is one of the disallowed values
	if password == "admin123" || password == "password123" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FailedResponse{
			ActionStatus:       "FAILED",
			FailureReason:      "common_password",
			FailureDescription: "The provided password is common password.",
		})
		return
	}

	// Success response if password is valid
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		ActionStatus: "SUCCESS",
	})
}

func main() {
	http.HandleFunc("/", passwordValidationHandler)

	port := 8080
	fmt.Printf("Server is running on port %d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
