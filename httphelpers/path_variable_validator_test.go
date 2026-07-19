package httphelpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetGuidFromRequestPath(t *testing.T) {
	testCases := []struct {
		name     string
		reqPath  string
		expected string
		ok       bool
	}{
		{"Valid GUID", "/service/85622399-b2b7-4e98-9a8d-628e28b9aeb4", "85622399-b2b7-4e98-9a8d-628e28b9aeb4", true},
		{"Invalid GUID", "/service/invalid-guid", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.reqPath, nil)
			id := strings.Split(tc.reqPath, "/")
			req.SetPathValue("id", id[2])
			guidVal, ok := GetGuidFromRequestPath("id", req)
			if guidVal != tc.expected || ok != tc.ok {
				t.Errorf("GetGuidFromRequestPath(%q) = (%q, %v), want (%q, %v)", tc.reqPath, guidVal, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestGetDateFromRequestPath(t *testing.T) {
	testCases := []struct {
		name     string
		reqPath  string
		expected time.Time
		ok       bool
	}{
		{"Valid Date", "/service/2023-10-05", time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC), true},
		{"Invalid Date", "/service/2023-10-32", time.Time{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.reqPath, nil)
			pathVars := strings.Split(tc.reqPath, "/")
			req.SetPathValue("startDate", pathVars[2])
			dateVal, ok := GetDateFromRequestPath("startDate", req)
			if !dateVal.Equal(tc.expected) || ok != tc.ok {
				t.Errorf("GetDateFromRequestPath(%q) = (%v, %v), want (%v, %v)", tc.reqPath, dateVal, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestGetIntFromRequestPath(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		pathVal string
		want    int
		wantOk  bool
	}{
		{
			name:    "Valid positive ID",
			varName: "id",
			pathVal: "123",
			want:    123,
			wantOk:  true,
		},
		{
			name:    "Zero ID",
			varName: "id",
			pathVal: "0",
			want:    0,
			wantOk:  false,
		},
		{
			name:    "Negative ID",
			varName: "id",
			pathVal: "-1",
			want:    0,
			wantOk:  false,
		},
		{
			name:    "Non-numeric ID",
			varName: "id",
			pathVal: "abc",
			want:    0,
			wantOk:  false,
		},
		{
			name:    "Empty ID",
			varName: "id",
			pathVal: "",
			want:    0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.SetPathValue(tt.varName, tt.pathVal)

			got, ok := GetIntFromRequestPath(tt.varName, req)
			if ok != tt.wantOk {
				t.Errorf("GetIntFromRequestPath() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if got != tt.want {
				t.Errorf("GetIntFromRequestPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntFromRequestPath_Override(t *testing.T) {
	SetPathValueLookup(func(r *http.Request, name string) string {
		if name == "id" {
			return "456"
		}
		return ""
	})
	t.Cleanup(func() { SetPathValueLookup(nil) })

	req := httptest.NewRequest(http.MethodGet, "/456", nil)

	got, ok := GetIntFromRequestPath("id", req)
	if !ok {
		t.Errorf("GetIntFromRequestPath() ok = false, want true")
	}
	if got != 456 {
		t.Errorf("GetIntFromRequestPath() got = %v, want 456", got)
	}
}

func TestGetGuidFromRequestPath_Override(t *testing.T) {
	testCases := []struct {
		name     string
		pathVal  string
		expected string
		ok       bool
	}{
		{"Valid GUID", "85622399-b2b7-4e98-9a8d-628e28b9aeb4", "85622399-b2b7-4e98-9a8d-628e28b9aeb4", true},
		{"Invalid GUID", "invalid-guid", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetPathValueLookup(func(r *http.Request, name string) string {
				if name == "id" {
					return tc.pathVal
				}
				return ""
			})
			t.Cleanup(func() { SetPathValueLookup(nil) })

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			guidVal, ok := GetGuidFromRequestPath("id", req)
			if guidVal != tc.expected || ok != tc.ok {
				t.Errorf("GetGuidFromRequestPath() = (%q, %v), want (%q, %v)", guidVal, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestGetDateFromRequestPath_Override(t *testing.T) {
	testCases := []struct {
		name     string
		pathVal  string
		expected time.Time
		ok       bool
	}{
		{"Valid Date", "2023-10-05", time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC), true},
		{"Invalid Date", "2023-10-32", time.Time{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetPathValueLookup(func(r *http.Request, name string) string {
				if name == "startDate" {
					return tc.pathVal
				}
				return ""
			})
			t.Cleanup(func() { SetPathValueLookup(nil) })

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			dateVal, ok := GetDateFromRequestPath("startDate", req)
			if !dateVal.Equal(tc.expected) || ok != tc.ok {
				t.Errorf("GetDateFromRequestPath() = (%v, %v), want (%v, %v)", dateVal, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestPathValueLookup_Race(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	done := make(chan bool)
	const iterations = 1000

	go func() {
		for range iterations {
			SetPathValueLookup(func(r *http.Request, name string) string {
				return "test"
			})
			SetPathValueLookup(nil)
		}
		done <- true
	}()

	go func() {
		for range iterations {
			_, _ = GetIntFromRequestPath("id", req)
			_, _ = GetGuidFromRequestPath("id", req)
			_, _ = GetDateFromRequestPath("id", req)
		}
		done <- true
	}()

	for range 2 {
		<-done
	}
}
