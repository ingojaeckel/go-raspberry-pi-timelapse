package main

import (
	"fmt"
	"runtime/debug"
)

// Set by the Makefile via -ldflags; see versionString for the fallback.
var gitCommit, builtAt string

// versionString prefers the values injected by the Makefile and falls back
// to the VCS information Go embeds in every build.
func versionString(commit, built string) string {
	if commit != "" && built != "" {
		return fmt.Sprintf("%s built at %s", commit, built)
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := vcsVersion(info.Settings); v != "" {
			return v
		}
	}
	return "1"
}

// vcsVersion formats the vcs.* build settings, or returns "" without a revision.
func vcsVersion(settings []debug.BuildSetting) string {
	vcs := map[string]string{}
	for _, s := range settings {
		vcs[s.Key] = s.Value
	}
	if vcs["vcs.revision"] == "" {
		return ""
	}
	dirty := ""
	if vcs["vcs.modified"] == "true" {
		dirty = "-dirty"
	}
	return fmt.Sprintf("%.7s%s built at %s", vcs["vcs.revision"], dirty, vcs["vcs.time"])
}
