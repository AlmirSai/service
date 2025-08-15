package logger

import (
	"context"
	"runtime/debug"
	"strconv"
	"strings"
)

// BuildInfo logs the Go build information of the current binary.
// Useful for debugging and version tracking in production.
func (log *Logger) BuildInfo(ctx context.Context) {
	var values []any

	// Retrieve build info from the compiled binary.
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// If build info is not available, log a warning and return.
		log.Warn(ctx, "build info not available")
		return
	}

	// Iterate over build settings and prepare key/value pairs for logging.
	for _, s := range info.Settings {
		key := s.Key
		if quoteKey(key) {
			// Quote keys that are empty or contain special characters.
			key = strconv.Quote(key)
		}

		value := s.Value
		if quoteValue(value) {
			// Quote values that contain spaces or special characters.
			value = strconv.Quote(value)
		}

		values = append(values, key, value)
	}

	// Add Go version and main module version explicitly.
	values = append(values, "goversion", info.GoVersion)
	values = append(values, "modversion", info.Main.Version)

	// Log the complete build info as structured fields.
	log.Info(ctx, "build info", values...)
}

// quoteKey determines whether the build setting key needs quoting.
// Keys are quoted if they are empty or contain '=', whitespace, or special characters.
func quoteKey(key string) bool {
	return len(key) == 0 || strings.ContainsAny(key, "= \t\r\n\"`")
}

// quoteValue determines whether the build setting value needs quoting.
// Values are quoted if they contain whitespace or special characters.
func quoteValue(value string) bool {
	return strings.ContainsAny(value, " \t\r\n\"`")
}
