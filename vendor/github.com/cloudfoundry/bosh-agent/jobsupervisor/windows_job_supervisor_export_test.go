// +build windows

package jobsupervisor

func SetPipeExePath(s string) (previous string) {
	previous = pipeExePath
	pipeExePath = s
	return previous
}

func GetPipeExePath() string {
	return pipeExePath
}
