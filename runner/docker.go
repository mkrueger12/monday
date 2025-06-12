package runner

import (
	"fmt"
	"os"
	"os/exec"
)

func RunCodexInDocker(image string, cliArgs []string, workdir string, envVars map[string]string) error {
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", workdir),
		"-w", "/workspace",
	}
	
	for k, v := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	
	args = append(args, image)
	args = append(args, cliArgs...)
	
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
