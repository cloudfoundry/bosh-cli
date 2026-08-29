//go:build !windows

package system

import (
	"strings"
)

// mergeEnv merges system and command environments variables.  Command variables
// override any system variable with the same key.
func mergeEnv(sysEnv []string, cmdEnv map[string]string) []string {
	var env []string
	// cmdEnv has precedence and overwrites any duplicate vars
	for k, v := range cmdEnv {
		env = append(env, k+"="+v)
	}
	for _, s := range sysEnv {
		if before, _, ok := strings.Cut(s, "="); ok {
			k := before // key
			if _, found := cmdEnv[k]; !found {
				env = append(env, s)
			}
		}
	}
	return env
}
