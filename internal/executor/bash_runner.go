package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/apex/log"
)

// BashRunner - runs shell commands.
type BashRunner struct {
	workingDir string
	Env        []string
}

// Env - global list of environment variables.
var Env []string

func stringHideSecrets(str string, secrets ...string) string {
	hiddenStr := str
	for _, s := range secrets {
		hiddenStr = strings.Replace(hiddenStr, s, "***", -1)
	}
	return hiddenStr
}

func (b *BashRunner) commandExecCommon(command string, outputBuff io.Writer, errBuff io.Writer) error {
	// Prepere command, set outputs, run.

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = outputBuff
	cmd.Stderr = errBuff

	if b.workingDir != "" {
		cmd.Dir = b.workingDir
	}
	// Add global environments.
	envTmp := append(os.Environ(), Env...)
	// Add environments of curent innstance.
	cmd.Env = append(envTmp, b.Env...)
	// Run command.
	err := cmd.Run()

	return err
}

// Run - exec command and hide secrets in log output.
func (b *BashRunner) Run(command string, secrets ...string) error {

	// Realtime command output for debug mode only.
	var output io.Writer
	output = os.Stdout

	// errOutput - error text.
	errOutput := &bytes.Buffer{}

	// Mask secrets with ***
	hiddenCommand := stringHideSecrets(command, secrets...)
	log.Debugf("Executing command \"%s\"", hiddenCommand)

	// Run command.
	err := b.commandExecCommon(command, output, errOutput)
	if err != nil {
		return fmt.Errorf("%s, output: \n%s", err.Error(), errOutput.String())
	}
	return nil
}

// RunMutely - exec command and hide secrets in output. Return command output and errors output.
func (b *BashRunner) RunMutely(command string, secrets ...string) (string, string, error) {
	output := &bytes.Buffer{}
	runerr := &bytes.Buffer{}
	// Mask secrets with ***
	hiddenCommand := stringHideSecrets(command, secrets...)
	log.Debugf("Executing command \"%s\"", hiddenCommand)
	err := b.commandExecCommon(command, output, runerr)
	return output.String(), runerr.String(), err
}

// NewBashRunner - create new bash runner.
func NewBashRunner(workingDir string, envVariables ...string) (*BashRunner, error) {
	// Create runner.
	runner := BashRunner{}
	fi, err := os.Stat(workingDir)
	if workingDir != "" {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %s does not exist", workingDir)
		}
		if !fi.Mode().IsDir() {
			return nil, fmt.Errorf("%s is not dir", workingDir)
		}
	}
	runner.workingDir = workingDir
	runner.Env = envVariables
	return &runner, nil
}
