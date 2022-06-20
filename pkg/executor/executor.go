package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/logging"
)

// ShRunner - runs shell commands.
type ShRunner struct {
	workingDir        string
	Env               []string
	Timeout           time.Duration
	LogLabels         []string
	ShowResultMessage bool
}

// Env - global list of environment variables.
var Env []string

// DefaultLogWriter default logging driver to write stdout.
var DefaultLogWriter io.Writer

// NewExecutor - create new sh runner.
func NewExecutor(workingDir string, envVariables ...string) (*ShRunner, error) {
	// Create runner.
	runner := ShRunner{}
	fi, err := os.Stat(workingDir)
	if workingDir != "" {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("executor: directory %s does not exist", workingDir)
		}
		if !fi.Mode().IsDir() {
			return nil, fmt.Errorf("executor: %s is not dir", workingDir)
		}
	}

	runner.workingDir = workingDir
	runner.Env = envVariables
	runner.Timeout = 0
	runner.ShowResultMessage = true
	return &runner, nil
}

func (b *ShRunner) commandExecCommon(outputBuff io.Writer, errBuff io.Writer, command string, args ...string) error {
	// Prepare command, set outputs, run.
	if config.Interrupted {
		return fmt.Errorf("interrupted")
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if b.Timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), b.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, command, args...)
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

	stopChan := make(chan struct{})
	sigChan := StartSigTrap(cmd, stopChan)
	defer sigChan.Close()
	cmd.Start()
	err := cmd.Wait()
	stopChan <- struct{}{}
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("executor: command timeout '%s'", command)
	}

	return err
}

func (b *ShRunner) commandExecCommonInShell(command string, outputBuff io.Writer, errBuff io.Writer) error {
	// Prepere command, set outputs, run.
	if config.Interrupted {
		return fmt.Errorf("interrupted")
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if b.Timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), b.Timeout)
		defer cancel()
	}
	// Add set -e to handle errors in multiline commands.
	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("set -e\n%v", command))
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
	stopChan := make(chan struct{})
	sigChan := StartSigTrap(cmd, stopChan)
	defer sigChan.Close()
	cmd.Start()
	err := cmd.Wait()
	stopChan <- struct{}{}
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("sh runner: command timeout '%s'", command)
	}

	return err
}

func (b *ShRunner) RunWithTty(command string) error {
	var ctx context.Context
	ctx = context.Background()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if b.workingDir != "" {
		cmd.Dir = b.workingDir
	}
	// Add global environments.
	log.Debugf("Run command '%s'", command)
	envTmp := append(os.Environ(), Env...)
	// Add environments of curent innstance.
	cmd.Env = append(envTmp, b.Env...)
	// Run command.
	err := cmd.Run()
	return err
}

// Run - exec command and return stdout, stderr, run error.
func (b *ShRunner) Run(command string) ([]byte, []byte, error) {

	var logPrefix string
	for _, str := range b.LogLabels {
		logPrefix = fmt.Sprintf("%s[%s]", logPrefix, str)
	}
	log.Infof("%s %-7s", logPrefix, colors.Fmt(colors.LightWhiteBold).Sprint("In progress..."))
	log.Debugf("%s Executing command '%s':", logPrefix, command)
	// Create log writer.
	logWriter, err := logging.NewLogWriter(log.DebugLevel, logging.SliceFielder{Flds: b.LogLabels})
	if err != nil {
		return nil, nil, fmt.Errorf("executor: can't init logging: %v", err)
	}

	// errOutput - error text.
	errOutput := &bytes.Buffer{}

	bannerStopChan := make(chan struct{})
	if config.Global.LogLevel != "debug" {

		// banner = fmt.Sprintf("%s[dir='%s'][cmd='%s']", banner, "./"+dir, command)
		banner := fmt.Sprintf("%s executing in progress...", logPrefix)
		go showBanner(banner, bannerStopChan)

		defer func(stop chan struct{}) {
			stop <- struct{}{}
			close(stop)
		}(bannerStopChan)
	}
	logCollector := newCollector(logWriter)
	err = b.commandExecCommonInShell(command, logCollector, errOutput)
	if b.ShowResultMessage {
		if err == nil {
			log.Infof("%s %-7s", logPrefix, colors.Fmt(colors.LightWhiteBold).Sprint("Success"))
		} else {
			log.Infof("%s %-7s", logPrefix, colors.Fmt(colors.LightWhiteBold).Sprint("Error"))
		}
	}
	return logCollector.Data(), errOutput.Bytes(), err
}

// RunMutely - exec command and hide secrets in output. Return command output and errors output.
func (b *ShRunner) RunMutely(command string, secrets ...string) (string, string, error) {
	var logPrefix string
	for _, str := range b.LogLabels {
		logPrefix = fmt.Sprintf("%s[%s]", logPrefix, str)
	}
	output := &bytes.Buffer{}
	runerr := &bytes.Buffer{}
	// Mask secrets with ***
	hiddenCommand := stringHideSecrets(command, secrets...)
	log.Debugf("Executing command '%s':", hiddenCommand)
	err := b.commandExecCommonInShell(command, output, runerr)
	return output.String(), runerr.String(), err
}

func stringHideSecrets(str string, secrets ...string) string {
	hiddenStr := str
	for _, s := range secrets {
		hiddenStr = strings.Replace(hiddenStr, s, "***", -1)
	}
	return hiddenStr
}

func showBanner(banner string, done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	start := time.Now()
	log.Infof("%s %-7s", banner, time.Since(start).Round(time.Second))
	for {
		select {
		case <-ticker.C:
			log.Infof("%s %-7s", banner, time.Since(start).Round(time.Second))
		case <-done:
			ticker.Stop()
			return
		}
	}
}

type SigTrap chan os.Signal

func StartSigTrap(cmd *exec.Cmd, stop chan struct{}) SigTrap {
	sChan := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	signal.Notify(sChan, signals...)
	go func() {
		for {
			select {
			case s := <-sChan:
				config.Interrupted = true
				config.Global.LogLevel = "debug"
				logging.InitLogLevel(config.Global.LogLevel, config.Global.TraceLog)
				log.Debugf("executor: forward signal %v", s)
				err := cmd.Process.Signal(s)
				if err != nil {
					log.Debugf("Error forwarding signal: %v", err)
				}
			case <-stop:
				return
			}
		}
	}()
	return sChan
}

func (signalChannel *SigTrap) Close() error {
	signal.Stop(*signalChannel)
	*signalChannel <- nil
	close(*signalChannel)
	return nil
}
