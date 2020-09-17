package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	json "github.com/json-iterator/go"

	"github.com/apex/log"
)

// TerraformRunner - exec terraform commands.
type TerraformRunner struct {
	workingDir string
	tfVarsJSON string
	bashRunner *BashRunner
	LogLabels  []string
}

// NewTerraformRunner create terraform runner.
func NewTerraformRunner(workingDir string, envVariables ...string) (*TerraformRunner, error) {
	var tr TerraformRunner
	var err error
	tr.bashRunner, err = NewBashRunner(workingDir, envVariables...)
	if err != nil {
		return nil, err
	}
	tr.workingDir = workingDir
	return &tr, nil
}

// SetLogLabels add log labels to hiden log output.
func (tr *TerraformRunner) SetLogLabels(logLabels []string) {
	tr.LogLabels = logLabels
}

func (tr *TerraformRunner) commonRun(command string, tfVars interface{}, tfBackend interface{}, args ...string) error {
	tfArgs := ""
	tr.bashRunner.LogLabels = tr.LogLabels
	// Create tfvars file in JSON forman from receiver tfVars interface{} struct.
	if tfVars != nil {
		// Convert struct to JSON.
		varsJSON, err := json.MarshalIndent(tfVars, "", " ")
		if err != nil {
			return err
		}
		// Create file.
		tfVarsFilename := filepath.Join(tr.workingDir, "vars-tmp.tfvars.json")
		// Write JSON in file.
		ioutil.WriteFile(tfVarsFilename, varsJSON, os.ModePerm)
		// Remove tmp file after func return.
		defer os.RemoveAll(tfVarsFilename)
		// Add var-file arg to command.
		tfArgs = "-var-file=vars-tmp.tfvars.json"
		log.Debugf("Terraform tfVars file: %s", string(varsJSON))
	}
	// Create tfvars file in JSON forman from receiver tfVars interface{} struct.
	if tfBackend != nil {
		// Convert struct to JSON.
		backendJSON, err := json.MarshalIndent(tfBackend, "", " ")
		if err != nil {
			return err
		}
		// Create file.
		backendFilename := filepath.Join(tr.workingDir, "backend-tmp.tfvars.json")
		// Write JSON in file.
		ioutil.WriteFile(backendFilename, backendJSON, os.ModePerm)
		// Remove tmp file after func return.
		defer os.RemoveAll(backendFilename)
		// Add var-file arg to command.
		tfArgs = "-backend-config=backend-tmp.tfvars.json"
		log.Debugf("Terraform backend config file: %s", string(backendJSON))
	}
	// Additional arguments.
	for _, arg := range args {
		tfArgs = fmt.Sprintf("%s %s", tfArgs, arg)
	}
	// Run command and return result.
	err := tr.bashRunner.Run(fmt.Sprintf("terraform %s %s", command, tfArgs))
	return err
}

// Version - exec terraform version.
func (tr *TerraformRunner) Version() (string, error) {
	var err error
	o, oerr, err := tr.bashRunner.RunMutely("terraform version")
	if err != nil {
		return "", fmt.Errorf("%s, error output: %s", err.Error(), oerr)
	}
	return o, nil
}

// Init - exec terraform init.
func (tr *TerraformRunner) Init(backendConfig interface{}) error {
	// Run command and return result.
	return tr.commonRun("init", nil, backendConfig)
}

// Plan - exec terraform plan.
func (tr *TerraformRunner) Plan(tfVars interface{}, args ...string) error {
	// Run command and return result.
	args = append(args, "-input=false")
	return tr.commonRun("plan", tfVars, nil, args...)
}

// Destroy - exec terraform apply.
func (tr *TerraformRunner) Destroy(tfVars interface{}, args ...string) error {
	// Run command and return result.
	args = append(args, "-input=false")
	args = append(args, "-auto-approve")
	return tr.commonRun("destroy", tfVars, nil, args...)
}

// Apply - exec terraform apply.
func (tr *TerraformRunner) Apply(tfVars interface{}, args ...string) error {
	// Run command and return result.
	args = append(args, "-input=false")
	args = append(args, "-auto-approve")
	return tr.commonRun("apply", tfVars, nil, args...)
}

// Import - exec terraform apply.
func (tr *TerraformRunner) Import(tfVars interface{}, args ...string) error {
	// Run command and return result.
	return tr.commonRun("import", tfVars, nil, args...)
}

// ApplyPlan - exec terraform apply.
func (tr *TerraformRunner) ApplyPlan(planFileName string, args ...string) error {
	// Run command and return result.
	args = append(args, "-input=false")
	args = append(args, "-auto-approve")
	args = append(args, planFileName)
	return tr.commonRun("apply", nil, nil, args...)
}

// Clear - remove .terraform
func (tr *TerraformRunner) Clear() error {
	// Run command and return result.
	err := os.RemoveAll(filepath.Join(tr.workingDir, ".terraform"))
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(tr.workingDir, "terraform.tfstate"))
}
