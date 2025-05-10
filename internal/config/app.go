package config

import "github.com/rs/zerolog/log"

var (
	ProjectID string
	Version   string
	GoVersion string
	BuildDate string
	GitLog    string
	GitHash   string
	GitBranch string
)

func LogProjetInfo() {
	log.Debug().
		Str("ProjectID", ProjectID).
		Str("Version", Version).
		Str("GoVersion", GoVersion).
		Str("BuildDate", BuildDate).
		Str("GitLog", GitLog).
		Str("GitHash", GitHash).
		Str("GitBranch", GitBranch)
}
