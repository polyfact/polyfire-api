package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type KeyValue struct {
	Key   string
	Value string
}

type (
	RecordFunc           func(string, ...KeyValue)
	RecordWithUserIDFunc func(string, string, ...KeyValue)
	RecordRequestFunc    func(string, string, string, ...KeyValue)
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

var ErrorMessages = map[string]APIError{
	// Authorization Errors
	"free_user_init_disabled": {
		Code:       "free_user_init_disabled",
		Message:    "Free user initialization is disabled for this project.",
		StatusCode: http.StatusForbidden,
	},
	"invalid_authorization_format": {
		Code:       "invalid_authorization_format",
		Message:    "Invalid format for Authorization header.",
		StatusCode: http.StatusUnauthorized,
	},
	"invalid_token": {Code: "invalid_token", Message: "Invalid token.", StatusCode: http.StatusUnauthorized},
	"missing_authorization": {
		Code:       "missing_authorization",
		Message:    "Authorization header is missing.",
		StatusCode: http.StatusUnauthorized,
	},
	"missing_user_id": {
		Code:       "missing_user_id",
		Message:    "User ID not found.",
		StatusCode: http.StatusUnauthorized,
	},
	"no_token": {Code: "no_token", Message: "No token provided.", StatusCode: http.StatusUnauthorized},
	"token_exchange_failed": {
		Code:       "token_exchange_failed",
		Message:    "Failed to exchange user token.",
		StatusCode: http.StatusUnauthorized,
	},
	"token_signature_error": {
		Code:       "token_signature_error",
		Message:    "Failed to sign the user token.",
		StatusCode: http.StatusUnauthorized,
	},
	"rate_limit_reached": {
		Code:       "rate_limit_reached",
		Message:    "You have reached the rate limit for this project",
		StatusCode: http.StatusTooManyRequests,
	},
	"project_rate_limit_reached": {
		Code:       "project_rate_limit_reached",
		Message:    "This project has reached its monthly rate limit",
		StatusCode: http.StatusTooManyRequests,
	},

	// Bad Request Errors (Communication, Content, Decoding, Generation, Model, and Methods)
	"communication_error": {
		Code:       "communication_error",
		Message:    "Failed to establish communication. Please try again later.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_content_type": {
		Code:       "invalid_content_type",
		Message:    "Expected 'application/json' content type.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_json": {
		Code:       "invalid_json",
		Message:    "Failed to decode request body. Ensure valid JSON format.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_message_type": {
		Code:       "invalid_message_type",
		Message:    "Invalid message type. Only text messages are supported.",
		StatusCode: http.StatusBadRequest,
	},
	"missing_content": {
		Code:       "missing_content",
		Message:    "The request is missing content.",
		StatusCode: http.StatusBadRequest,
	},
	"missing_id": {
		Code:       "missing_id",
		Message:    "Missing ID parameter in the request.",
		StatusCode: http.StatusBadRequest,
	},
	"decode_error": {
		Code:       "decode_error",
		Message:    "Failed to decode the incoming request. Please verify the request format.",
		StatusCode: http.StatusBadRequest,
	},
	"generation_error": {
		Code:       "generation_error",
		Message:    "An error occurred while starting the generation. Please try again.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_model_provider": {
		Code:       "invalid_model_provider",
		Message:    "Provided model provider is unknown.",
		StatusCode: http.StatusBadRequest,
	},
	"only_post_method_allowed": {
		Code:       "only_post_method_allowed",
		Message:    "Only POST method is allowed for this endpoint.",
		StatusCode: http.StatusMethodNotAllowed,
	},
	"read_message_error": {
		Code:       "read_message_error",
		Message:    "Failed to read the message. Please ensure the request is valid.",
		StatusCode: http.StatusBadRequest,
	},
	"unknown_model_provider": {
		Code:       "unknown_model_provider",
		Message:    "The provided model is unknown or unsupported.",
		StatusCode: http.StatusBadRequest,
	},

	// Database Errors
	"database_error": {
		Code:       "database_error",
		Message:    "Failed to set the key-value pair in the database.",
		StatusCode: http.StatusInternalServerError,
	},
	"db_creation_error": {
		Code:       "db_creation_error",
		Message:    "Failed to create memory in the database.",
		StatusCode: http.StatusInternalServerError,
	},
	"db_insert_error": {
		Code:       "db_addition_error",
		Message:    "Failed to add memory to the database.",
		StatusCode: http.StatusInternalServerError,
	},
	"retrieval_error": {
		Code:       "retrieval_error",
		Message:    "Failed to retrieve memory IDs from the database.",
		StatusCode: http.StatusInternalServerError,
	},
	"storage_error": {
		Code:       "storage_error",
		Message:    "Failed to store a fail in the bucket",
		StatusCode: http.StatusInternalServerError,
	},

	// Prompt Errors
	"decode_prompt_error": {
		Code:       "decode_prompt_error",
		Message:    "Failed to decode the prompt data.",
		StatusCode: http.StatusInternalServerError,
	},
	"db_fetch_prompt_error": {
		Code:       "db_fetch_prompt_error",
		Message:    "Failed to fetch prompt from the database.",
		StatusCode: http.StatusNotFound,
	},
	"db_insert_prompt_error": {
		Code:       "db_insert_prompt_error",
		Message:    "Failed to insert prompt into the database.",
		StatusCode: http.StatusBadRequest,
	},
	"db_update_prompt_error": {
		Code:       "db_update_prompt_error",
		Message:    "Failed to update the prompt in the database.",
		StatusCode: http.StatusBadRequest,
	},
	"db_delete_prompt_error": {
		Code:       "db_delete_prompt_error",
		Message:    "Failed to delete the prompt from the database.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_filter_operation": {
		Code:       "invalid_filter_operation",
		Message:    "Invalid filter operation.",
		StatusCode: http.StatusBadRequest,
	},
	"invalid_column": {
		Code:       "invalid_column",
		Message:    "Filter column name not allowed",
		StatusCode: http.StatusForbidden,
	},
	"invalid_length_value": {
		Code:       "invalid_length_value",
		Message:    "Invalid length value.",
		StatusCode: http.StatusBadRequest,
	},

	// Embedding Errors
	"embedding_error": {
		Code:       "embedding_error",
		Message:    "Failed to process the embedding.",
		StatusCode: http.StatusInternalServerError,
	},

	// Not Found Errors
	"data_not_found": {Code: "data_not_found", Message: "The requested data was not found.", StatusCode: http.StatusNotFound},
	"not_found":      {Code: "not_found", Message: "Requested resource not found.", StatusCode: http.StatusNotFound},
	"not_found_error": {
		Code:       "not_found_error",
		Message:    "Resource not found. Please check your request.",
		StatusCode: http.StatusNotFound,
	},
	"user_id_error": {Code: "user_id_error", Message: "User ID not found in context.", StatusCode: http.StatusNotFound},

	// Project Errors
	"project_retrieval_error": {
		Code:       "project_retrieval_error",
		Message:    "There was an error retrieving the project. The projectId is probably invalid.",
		StatusCode: http.StatusInternalServerError,
	},
	"project_user_creation_failed": {
		Code:       "project_user_creation_failed",
		Message:    "Failed to create a user for the project.",
		StatusCode: http.StatusInternalServerError,
	},

	// Internal Server Errors (Chat, Writing, Reading, Transcription)
	"error_chat_history": {
		Code:       "error_chat_history",
		Message:    "Failed to retrieve chat history. Please try again later.",
		StatusCode: http.StatusInternalServerError,
	},
	"error_create_chat": {
		Code:       "error_create_chat",
		Message:    "Failed to create the chat. Please try again later.",
		StatusCode: http.StatusInternalServerError,
	},
	"internal_error": {
		Code:       "internal_error",
		Message:    "An internal error occurred. Please try again later.",
		StatusCode: http.StatusInternalServerError,
	},
	"read_error": {
		Code:       "read_error",
		Message:    "Failed to read the request content.",
		StatusCode: http.StatusInternalServerError,
	},
	"splitting_error": {
		Code:       "splitting_error",
		Message:    "Failed to split the input text.",
		StatusCode: http.StatusInternalServerError,
	},
	"transcription_error": {
		Code:       "transcription_error",
		Message:    "Failed to transcribe the audio.",
		StatusCode: http.StatusInternalServerError,
	},
	"image_generation_error": {
		Code:       "image_generation_error",
		Message:    "Failed to generate this image.",
		StatusCode: http.StatusInternalServerError,
	},
	"write_end_message_error": {
		Code:       "write_end_message_error",
		Message:    "Failed to write the end message to connection.",
		StatusCode: http.StatusInternalServerError,
	},
	"write_info_error": {
		Code:       "write_info_error",
		Message:    "Failed to write info message to connection.",
		StatusCode: http.StatusInternalServerError,
	},
	"write_message_error": {
		Code:       "write_message_error",
		Message:    "Failed to write message to connection.",
		StatusCode: http.StatusInternalServerError,
	},

	// Web request

	"error_website_exceeds_limit": {
		Code:       "error_website_exceeds_limit",
		Message:    "Error: The token count for a website exceeds the allowed limit.",
		StatusCode: http.StatusBadRequest,
	},
	"error_websites_content_exceeds": {
		Code:       "error_websites_content_exceeds",
		Message:    "Error: The accumulated content from the websites exceeds the token limit.",
		StatusCode: http.StatusBadRequest,
	},
	"error_no_content_found": {
		Code:       "error_no_content_found",
		Message:    "Error: No content accumulated from the results.",
		StatusCode: http.StatusBadRequest,
	},
	"error_fetch_webpage": {
		Code:       "error_fetch_webpage",
		Message:    "Error fetching the webpage:",
		StatusCode: http.StatusInternalServerError,
	},
	"error_parse_content": {
		Code:       "error_parse_content",
		Message:    "Error parsing the webpage content with readability lib.",
		StatusCode: http.StatusInternalServerError,
	},
	"error_visit_base_url": {
		Code:       "error_visit_base_url",
		Message:    "Error visiting the search engine URL",
		StatusCode: http.StatusInternalServerError,
	},

	// Fallback error

	"unknown_error": {Code: "unknown_error", Message: "An unknown error occurred.", StatusCode: http.StatusInternalServerError},
}

func RespondError(w http.ResponseWriter, record RecordFunc, errorCode string, message ...string) {
	apiError, exists := ErrorMessages[errorCode]

	if !exists {
		apiError = ErrorMessages["unknown_error"]
	}

	if len(message) > 0 {
		apiError.Message = message[0]
	}

	log.Println(apiError)
	error_bytes, _ := json.Marshal(&apiError)
	record(string(error_bytes), KeyValue{Key: "Error", Value: "true"})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.StatusCode)
	json.NewEncoder(w).Encode(apiError)
}

func RespondErrorStream(conn *websocket.Conn, record RecordFunc, errorCode string, message ...string) {
	apiError, exists := ErrorMessages[errorCode]

	if !exists {
		apiError = ErrorMessages["unknown_error"]
	}

	if len(message) > 0 {
		apiError.Message = message[0]
	}

	log.Println(apiError)
	error_bytes, _ := json.Marshal(&apiError)
	record(string(error_bytes), KeyValue{Key: "Error", Value: "true"})

	res, _ := json.Marshal(apiError)

	conn.WriteMessage(websocket.TextMessage, []byte("[ERROR]:"+string(res)))
}
