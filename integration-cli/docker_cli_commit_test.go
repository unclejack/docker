package main

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestCommitAfterContainerIsDone(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-i", "-a", "stdin", "busybox", "echo", "foo")
	out, _, err := runCommandWithOutput(runCmd)
	errorOut(err, t, fmt.Sprintf("failed to run container: %v %v", out, err))

	cleanedContainerID := stripTrailingCharacters(out)

	commitCmd := exec.Command(dockerBinary, "commit", cleanedContainerID)
	out, _, err = runCommandWithOutput(commitCmd)
	errorOut(err, t, fmt.Sprintf("failed to commit container to image: %v %v", out, err))

	cleanedImageID := stripTrailingCharacters(out)

	inspectCmd := exec.Command(dockerBinary, "inspect", cleanedImageID)
	out, _, err = runCommandWithOutput(inspectCmd)
	errorOut(err, t, fmt.Sprintf("failed to inspect image: %v %v", out, err))

	go deleteContainer(cleanedContainerID)
	go deleteImages(cleanedImageID)

	logDone("commit - echo foo and commit the image")
}
