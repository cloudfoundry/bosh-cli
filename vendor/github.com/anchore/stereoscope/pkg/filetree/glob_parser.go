package filetree

import (
	"regexp"
	"strings"
)

const (
	// searchByGlob is the default, unparsed/processed glob value searched directly against the filetree.
	searchByGlob searchBasis = iota

	// searchByFullPath indicates that the given glob value is not a glob, thus a (simpler) path lookup against the filetree should be performed as the search.
	searchByFullPath

	// searchByExtension indicates cases like "**/*.py" where the only specific glob element indicates the file or directory extension.
	searchByExtension

	// searchByBasename indicates cases like "**/bin/python" where the only specific glob element indicates the file or directory basename (e.g. "python").
	searchByBasename

	// searchByBasenameGlob indicates cases like "**/bin/python*" where the search space is limited to the full set of all basenames that match the given glob.
	searchByBasenameGlob

	// searchBySubDirectory indicates cases like "**/var/lib/dpkg/status.d/*" where we're interested in selecting all files within a directory (but not the directory itself).
	searchBySubDirectory
)

type searchBasis int

func (s searchBasis) String() string {
	switch s {
	case searchByGlob:
		return "glob"
	case searchByFullPath:
		return "full-path"
	case searchByExtension:
		return "extension"
	case searchByBasename:
		return "basename"
	case searchByBasenameGlob:
		return "basename-glob"
	case searchBySubDirectory:
		return "subdirectory"
	}
	return "unknown search basis"
}

type searchRequest struct {
	searchBasis
	value       string
	requirement string
}

func (s searchRequest) String() string {
	value := s.searchBasis.String() + ": " + s.value
	if s.requirement != "" {
		value += " (requirement: " + s.requirement + ")"
	}
	return value
}

func parseGlob(glob string) []searchRequest {
	glob = cleanGlob(glob)

	if !strings.ContainsAny(glob, "*?[]{}") {
		return []searchRequest{
			{
				searchBasis: searchByFullPath,
				value:       glob,
			},
		}
	}

	beforeBasename, basename := splitAtBasename(glob)

	if basename == "*" {
		_, nestedBasename := splitAtBasename(beforeBasename)
		if !strings.ContainsAny(nestedBasename, "*?[]{}") {
			// special case: glob is a parent glob
			requests := []searchRequest{
				{
					searchBasis: searchBySubDirectory,
					value:       nestedBasename,
					requirement: beforeBasename,
				},
			}
			return requests
		}
	}

	requests := parseGlobBasename(basename)
	for i := range requests {
		applyRequirement(&requests[i], beforeBasename, glob)
	}

	return requests
}

func splitAtBasename(glob string) (string, string) {
	// TODO: need to correctly avoid indexes within [] and {} groups
	basenameSplitAt := strings.LastIndex(glob, "/")

	var basename string
	var beforeBasename string
	if basenameSplitAt == -1 {
		// note: this has no glob path prefix, thus no requirement...
		// this can only be a basename, basename glob, or extension
		basename = glob
		beforeBasename = ""
	} else if basenameSplitAt < len(glob)-1 {
		basename = glob[basenameSplitAt+1:]
	}

	if basenameSplitAt >= 0 && basenameSplitAt < len(glob)-1 {
		beforeBasename = glob[:basenameSplitAt]
	}

	return beforeBasename, basename
}

func applyRequirement(request *searchRequest, beforeBasename, glob string) {
	var requirement string

	if beforeBasename != "" {
		requirement = glob
		switch beforeBasename {
		case "**", request.requirement:
			if request.searchBasis != searchByExtension {
				requirement = ""
			}
		}
	} else {
		requirement = ""
	}

	request.requirement = requirement

	if request.searchBasis == searchByGlob {
		request.value = glob
		if glob == request.requirement {
			request.requirement = ""
		}
	}
}

func parseGlobBasename(basenameInput string) []searchRequest {
	if strings.ContainsAny(basenameInput, "[]{}") {
		return parseBasenameAltAndClassGlobSections(basenameInput)
	}

	extensionFields := strings.Split(basenameInput, "*.")
	if len(extensionFields) == 2 && extensionFields[0] == "" {
		possibleExtension := extensionFields[1]
		if !strings.ContainsAny(possibleExtension, "*?") {
			// special case, this is plain extension
			return []searchRequest{
				{
					searchBasis: searchByExtension,
					value:       "." + possibleExtension,
				},
			}
		}
	}

	if !strings.ContainsAny(basenameInput, "*?") {
		// special case, this is plain basename
		return []searchRequest{
			{
				searchBasis: searchByBasename,
				value:       basenameInput,
			},
		}
	}

	if strings.ReplaceAll(strings.ReplaceAll(basenameInput, "?", ""), "*", "") == "" {
		// special case, this is a glob that is only asterisks... do not process!
		return []searchRequest{
			{
				searchBasis: searchByGlob,
				// note: we let the parent caller attach the full glob value
			},
		}
	}

	return []searchRequest{
		{
			searchBasis: searchByBasenameGlob,
			value:       basenameInput,
		},
	}
}

func parseBasenameAltAndClassGlobSections(basenameInput string) []searchRequest {
	// TODO: process escape sequences

	altStartCount := strings.Count(basenameInput, "{")
	altEndCount := strings.Count(basenameInput, "}")
	classStartCount := strings.Count(basenameInput, "[")
	classEndCount := strings.Count(basenameInput, "]")

	if altStartCount != altEndCount || classStartCount != classEndCount {
		// imbalanced braces, this is not a valid glob relative to just the basename
		return []searchRequest{
			{
				searchBasis: searchByGlob,
				// note: we let the parent caller attach the full glob value
			},
		}
	}

	if classStartCount > 0 {
		// parsing this is not supported at this time
		return []searchRequest{
			{
				searchBasis: searchByBasenameGlob,
				value:       basenameInput,
			},
		}
	}

	// if the glob is the simplest list form, them allow for breaking into sub-searches
	if altStartCount == 1 {
		indexStartIsPrefix := strings.Index(basenameInput, "{") == 0
		indexEndIsSuffix := strings.Index(basenameInput, "}") == len(basenameInput)-1
		if indexStartIsPrefix && indexEndIsSuffix {
			// this is a simple list, split it up
			// e.g. {a,b,c} -> a, b, c
			altSections := strings.Split(basenameInput[1:len(basenameInput)-1], ",")
			if len(altSections) > 1 {
				var requests []searchRequest
				for _, altSection := range altSections {
					basis := searchByBasename
					if strings.ContainsAny(altSection, "*?") {
						basis = searchByBasenameGlob
					}

					requests = append(requests, searchRequest{
						searchBasis: basis,
						value:       altSection,
					})
				}
				return requests
			}
		}
	}

	// there is some sort of alt usage, but it is not a simple list... just treat it as a glob
	return []searchRequest{
		{
			searchBasis: searchByBasenameGlob,
			value:       basenameInput,
		},
	}
}

func cleanGlob(glob string) string {
	glob = strings.TrimSpace(glob)
	glob = removeRedundantCountGlob(glob, '/', 1)
	glob = removeRedundantCountGlob(glob, '*', 2)
	if len(glob) > 1 {
		// input case: /
		// then preserve the slash
		glob = strings.TrimRight(glob, "/")
	}
	// e.g. replace "/bar**/" with "/bar*/"
	glob = simplifyMultipleGlobAsterisks(glob)
	glob = simplifyGlobRecursion(glob)
	return glob
}

func simplifyMultipleGlobAsterisks(glob string) string {
	// this will replace any recursive globs (**) that are not clearly indicating recursive tree searches with a single *

	var sb strings.Builder
	var asteriskBuff strings.Builder
	var withinRecursiveStreak bool

	for idx, c := range glob {
		isAsterisk := c == '*'
		isSlash := c == '/'

		// special case, this is the first character in the glob and it is an asterisk...
		// treat this like a recursive streak
		if idx == 0 && isAsterisk {
			withinRecursiveStreak = true
			asteriskBuff.WriteRune(c)
			continue
		}

		if isAsterisk {
			asteriskBuff.WriteRune(c)
			continue
		}

		if isSlash {
			if withinRecursiveStreak {
				// this is a confirmed recursive streak
				// keep all asterisks!
				sb.WriteString(asteriskBuff.String())
				asteriskBuff.Reset()
			}

			if asteriskBuff.Len() > 0 {
				// this is NOT a recursive streak, but there are asterisks
				// keep only one asterisk
				sb.WriteRune('*')
				asteriskBuff.Reset()
			}

			// this is potentially a new streak...
			withinRecursiveStreak = true
		} else {
			// ... and this is NOT a recursive streak
			if asteriskBuff.Len() > 0 {
				// ... keep only one asterisk, since it's not recursive
				sb.WriteRune('*')
			}
			asteriskBuff.Reset()
			withinRecursiveStreak = false
		}

		sb.WriteRune(c)
	}

	if asteriskBuff.Len() > 0 {
		if withinRecursiveStreak {
			sb.WriteString(asteriskBuff.String())
		} else {
			sb.WriteRune('*')
		}
	}

	return sb.String()
}

var globRecursionRightPattern = regexp.MustCompile(`(\*\*/?)+`)

func simplifyGlobRecursion(glob string) string {
	// this function assumes that all redundant asterisks have been removed (e.g. /****/ -> /**/)
	// and that all seemingly recursive globs have been replaced with a single asterisk (e.g. /bar**/ -> /bar*/)
	glob = globRecursionRightPattern.ReplaceAllString(glob, "**/")
	glob = strings.ReplaceAll(glob, "//", "/")
	if strings.HasPrefix(glob, "/**/") {
		glob = strings.TrimPrefix(glob, "/")
	}
	if len(glob) > 1 {
		// input case: /**
		// then preserve the slash
		glob = strings.TrimRight(glob, "/")
	}
	return glob
}

func removeRedundantCountGlob(glob string, val rune, count int) string {
	var sb strings.Builder

	var streak int
	for _, c := range glob {
		if c == val {
			streak++
			if streak > count {
				continue
			}
		} else {
			streak = 0
		}

		sb.WriteRune(c)
	}
	return sb.String()
}
