package utils

import "fmt"

// CORSOrigin returns the allowed CORS origin for the given environment.
// frontEndClient is the front-end host (config.FrontEndClient).
func CORSOrigin(env, frontEndClient string) (string, error) {
	switch env {
	case "local", "development":
		return frontEndClient, nil
	case "production":
		return frontEndClient, nil
	default:
		return "", fmt.Errorf("cannot load origins for CORS policy: unknown env %q", env)
	}
}
