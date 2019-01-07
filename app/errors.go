package app

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONError struct {
	Message string `json:"message"`
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		if err, ok := msgAndArgs[0].(error); ok {
			return err.Error()
		}
		return msgAndArgs[0].(string)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}

func WriteJSONError(w http.ResponseWriter, code int, msgAndArgs ...interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&JSONError{messageFromMsgAndArgs(msgAndArgs...)})
}

func WriteBadRequestError(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusBadRequest, msgAndArgs...)
}

func WriteInternalServerError(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusInternalServerError, msgAndArgs...)
}

func WriteForbiddenError(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusForbidden, msgAndArgs...)
}

func WriteUnauthorizedError(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusUnauthorized, msgAndArgs...)
}

func WriteNotFoundError(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusNotFound, msgAndArgs...)
}

func WriteUnprocessableEntity(w http.ResponseWriter, msgAndArgs ...interface{}) {
	WriteJSONError(w, http.StatusUnprocessableEntity, msgAndArgs...)
}
