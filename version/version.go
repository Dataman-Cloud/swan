package version

import (
	"html/template"
	"io"
	"runtime"
)

// set by build LD_FLAGS
var (
	version   string
	buildTime string
	gitCommit string
)

type Version struct {
	Version   string `json:"version"`
	GitCommit string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

func GetVersion() Version {
	return Version{
		Version:   version,
		GitCommit: gitCommit,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

var versionTemplate = ` Version:      {{.Version}}
 Git commit:   {{.GitCommit}}
 Go version:   {{.GoVersion}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}
`

func TextFormatTo(w io.Writer) error {
	tmpl, _ := template.New("version").Parse(versionTemplate)
	return tmpl.Execute(w, GetVersion())
}
