package internal

import (
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func FindLatestVersion(current *semver.Version, tags []string, major, minor, patch bool) string {
	if major {
		minor = true
		patch = true
	}
	if minor {
		patch = true
	}

	currentIsStrict := len(strings.SplitN(current.Original(), ".", 3)) == 3

	type VersionTag struct {
		Version *semver.Version
		Tag     string
	}
	var versionTags []VersionTag

	// Collect valid semantic versions
	for _, tag := range tags {
		// Attempt to parse the tag as a semantic version to compare it later easily
		var v *semver.Version
		var err error
		if currentIsStrict {
			// Don't regress from a strict version to a non-strict one
			v, err = semver.StrictNewVersion(tag)
		} else {
			v, err = semver.NewVersion(tag)
		}
		if err != nil {
			continue
		}

		versionTags = append(versionTags, VersionTag{Version: v, Tag: tag})
	}

	if len(versionTags) == 0 {
		return ""
	}

	// Sort versions in descending order
	// This is necessary to find the latest version
	sort.Slice(versionTags, func(i, j int) bool {
		return versionTags[i].Version.GreaterThan(versionTags[j].Version)
	})

	for _, vt := range versionTags {
		v := vt.Version
		tag := vt.Tag

		// Skip versions not newer than current
		if v.LessThanEqual(current) || v.Prerelease() != current.Prerelease() {
			continue
		}

		accept := false
		if major && v.Major() > current.Major() {
			accept = true
		} else if minor && isEqualMajor(v, current) && v.Minor() > current.Minor() {
			accept = true
		} else if patch && isEqualMajor(v, current) && isEqualMinor(v, current) && v.Patch() > current.Patch() {
			accept = true
		}

		if accept {
			return tag
		}
	}

	return ""
}

func isEqualMajor(current, tag *semver.Version) bool {
	return current.Major() == tag.Major()
}

func isEqualMinor(current, tag *semver.Version) bool {
	return current.Minor() == tag.Minor()
}
