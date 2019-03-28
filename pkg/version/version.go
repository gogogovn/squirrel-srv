package version

var (
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of current build
	Release = "unset"
)

// Version shows API version info
type Version struct {
	BuildTime string `json:"buildTime"`
	Commit    string `json:"commit"`
	Release   string `json:"release"`
}

// Info shows version info
func Info() *Version {
	return &Version{BuildTime, Commit, Release}
}
