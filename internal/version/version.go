package version

import (
	"debug/buildinfo"
	"runtime/debug"
)

func GetRevision(path string) string {
	var info *debug.BuildInfo

	if path == "" {
		var ok bool
		if info, ok = debug.ReadBuildInfo(); !ok {
			return ""
		}
	} else {
		var err error
		info, err = buildinfo.ReadFile(path)
		if err != nil {
			return ""
		}
	}

	var (
		vcs      string
		revision string
		modified string
	)

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs":
			vcs = setting.Value
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				modified = "-modified"
			}
		}
	}

	return vcs + "-" + revision + modified
}

func GetVersion(path string) string {
	var info *debug.BuildInfo

	if path == "" {
		var ok bool
		if info, ok = debug.ReadBuildInfo(); !ok {
			return ""
		}
	} else {
		var err error
		info, err = buildinfo.ReadFile(path)
		if err != nil {
			return ""
		}
	}

	return info.Main.Version
}
