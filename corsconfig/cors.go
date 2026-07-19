package corsconfig

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"slices"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

func getEnvVarValue(key string) string {
	val, found := os.LookupEnv(key)
	if !found {
		log.Println("Environment variable " + key + " not found")
	}
	return val
}

func getDefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
	}
}

func GetCORSConfig() CORSConfig {
	configStr := getEnvVarValue("CORS_CONFIG")
	slog.Debug("CORS config: ", slog.String("config", configStr))
	if configStr == "" {
		return getDefaultCORSConfig()
	}
	corsConfig := &CORSConfig{}
	err := json.Unmarshal([]byte(configStr), corsConfig)
	if err != nil {
		slog.Warn("Error parsing CORS config: ", slog.Any("error", err))
		return getDefaultCORSConfig()
	}

	// Safety check: if AllowedOrigins contains wildcard, we must force AllowCredentials to false
	if slices.Contains(corsConfig.AllowedOrigins, "*") {
		corsConfig.AllowCredentials = false
	}

	return *corsConfig
}
