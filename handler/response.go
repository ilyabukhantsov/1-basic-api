// Package handlers contains HTTP handlers for the API.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"1-basic-api/database"
)

const maxBodyBytes = 1 << 20 // 1 MiB

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON reads a JSON body into v, rejecting unknown fields and oversized bodies.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return false
		}
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	if dec.Decode(&struct{}{}) != io.EOF {
		writeError(w, http.StatusBadRequest, "Request body must contain a single JSON object")
		return false
	}
	return true
}

// pathID parses the named path parameter as a positive integer ID.
func pathID(w http.ResponseWriter, r *http.Request, name string) (uint, bool) {
	id, err := strconv.ParseUint(r.PathValue(name), 10, 32)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, "Invalid "+name)
		return 0, false
	}
	return uint(id), true
}

// writeDBError maps database errors to HTTP responses.
func writeDBError(w http.ResponseWriter, err error, entity string) {
	switch {
	case errors.Is(err, database.ErrNotFound):
		writeError(w, http.StatusNotFound, entity+" not found: "+err.Error())
	case errors.Is(err, database.ErrConflict):
		writeError(w, http.StatusConflict, entity+" already exists")
	default:
		log.Printf("database error: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
	}
}
