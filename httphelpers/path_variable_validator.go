package httphelpers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	pathValueLookup   func(*http.Request, string) string
	pathValueLookupMu sync.RWMutex
)

// SetPathValueLookup sets a custom function to retrieve values from request paths based on a key.
func SetPathValueLookup(fn func(*http.Request, string) string) {
	pathValueLookupMu.Lock()
	defer pathValueLookupMu.Unlock()
	pathValueLookup = fn
}

type PathValidator func(string, *http.Request) (string, bool)

func getPathValue(req *http.Request, varName string) string {
	val := req.PathValue(varName)
	if val != "" {
		return val
	}
	pathValueLookupMu.RLock()
	defer pathValueLookupMu.RUnlock()
	if pathValueLookup != nil {
		return pathValueLookup(req, varName)
	}
	return ""
}

func GetGuidFromRequestPath(varName string, req *http.Request) (string, bool) {
	guidVal := getPathValue(req, varName)
	return IsValidGuid(guidVal)
}

func IsValidGuid(guidVal string) (string, bool) {
	err := uuid.Validate(guidVal)
	if err != nil {
		return "", false
	}
	return guidVal, true
}

func GetDateFromRequestPath(varName string, req *http.Request) (time.Time, bool) {
	dateVal := getPathValue(req, varName)
	date, err := time.Parse("2006-01-02", dateVal)
	return date, err == nil
}

func GetIntFromRequestPath(varName string, req *http.Request) (int, bool) {
	val := getPathValue(req, varName)
	if val == "" {
		return 0, false
	}
	id, err := strconv.Atoi(val)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
