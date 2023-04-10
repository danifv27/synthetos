package version

import "fmt"

// Version is responsible for exposing version operations to the application layer
type Version interface {
	//GetVersionInfo
	GetVersionInfo() (VersionInfo, error)
}

type VersionInfo struct {
	//GitCommit The git commit that was compiled. This will be filled in by the compiler.
	GitCommit string
	//Version The main version number that is being run at the moment.
	Version  string
	Revision string
	Branch   string
	//BuildDate This will be filled in by the makefile
	BuildDate string
	BuildUser string
	//GoVersion The runtime version
	GoVersion string
	//OsArch The OS architecture
	OsArch string
	//Application Name
	Name string
}

func (v VersionInfo) String() string {

	return fmt.Sprintf("Version:\t%s\nGit commit:\t%s\nBuilt:\t\t%s (from %s by %s)",
		v.Version, v.GitCommit, v.BuildDate, v.Branch, v.BuildUser)
}
