package embed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"

	aversion "fry.org/cmo/cli/internal/application/version"
)

var (
	//embed doesn't allow cross package boundaries, so version.json should be in this folder
	//go:embed version.json
	version string
	//Application Name
	Name string
)

type Version struct {
	versioninfo aversion.VersionInfo
}

// NewVersionRepo Constructor
func NewVersion() aversion.Version {
	var j map[string]interface{}

	if err := json.Unmarshal([]byte(version), &j); err != nil {
		return &Version{}
	}

	v := Version{}
	// The value of your map associated with key "git" is of type map[string]interface{}.
	// And we want to access the element of that map associated with the key "commit".
	// .(string) type assertion to convert interface{} to string
	v.versioninfo.GitCommit = j["git"].(map[string]interface{})["commit"].(string)
	v.versioninfo.Branch = j["git"].(map[string]interface{})["branch"].(string)
	v.versioninfo.Version = j["version"].(string)
	v.versioninfo.Revision = j["revision"].(string)
	v.versioninfo.BuildDate = j["build"].(map[string]interface{})["date"].(string)
	v.versioninfo.BuildUser = j["build"].(map[string]interface{})["user"].(string)
	v.versioninfo.GoVersion = runtime.Version()
	v.versioninfo.OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	// v.versioninfo.Name = Name

	return &v
}

// GetVersionInfo Returns the version information
func (m *Version) GetVersionInfo() (aversion.VersionInfo, error) {

	return m.versioninfo, nil
}
