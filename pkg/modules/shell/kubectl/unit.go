package base

import (
	"fmt"
	"path/filepath"

	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

type YAMLFilesList struct {
	Files common.FilesListT
}

type KubectlCliT struct {
	Version     string `yaml:"recursive" json:"recursive"`
	Autoinstall bool   `yaml:"autoinstall" json:"autoinstall"`
}

type KustomizeT struct {
	Path string `yaml:"path" json:"path"`
}

// Unit describe cluster.dev unit to deploy/destroy k8s resources with kubectl.
type Unit struct {
	common.Unit
	Namespace      string             `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Kubeconfig     string             `yaml:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	KubectlOpts    string             `yaml:"kubectl_opts,omitempty" json:"kubectl_opts,omitempty"`
	KubectlCliConf *KubectlCliT       `yaml:"kubectl,omitempty" json:"kubectl,omitempty"`
	Path           string             `yaml:"path" json:"path"`
	ManifestsFiles *common.FilesListT `yaml:"-" json:"manifests"`
	ApplyTemplate  bool               `yaml:"apply_template" json:"-"`
	recursive      bool               `yaml:"-" json:"-"`
	UnitKind       string             `yaml:"-" json:"type"`
	SavedState     interface{}        `yaml:"-" json:"-"`
}

var kubectlBin = "kubectl"

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) fillShellUnit() {
	commandOpts := ""
	if u.recursive {
		commandOpts += "-R "
	}
	if u.Namespace != "" {
		commandOpts += "-n " + u.Namespace
	}
	if u.KubectlOpts != "" {
		commandOpts = fmt.Sprintf("%s %s", commandOpts, u.KubectlOpts)
	}
	if u.Kubeconfig != "" {
		commandOpts = fmt.Sprintf("%s --kubeconfig='%s'", commandOpts, u.Kubeconfig)
	}
	u.ApplyConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s apply %s -f %s", kubectlBin, commandOpts, filepath.Join(u.CacheDir, u.Path)),
		},
	}
	// log.Warnf("path: %v", u.Path)
	u.DestroyConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s delete %s -f %s", kubectlBin, commandOpts, filepath.Join(u.CacheDir, u.Path)),
		},
	}
	// u.CreateFiles = nil
	// log.Warnf("apply: %+v", u.ApplyConf)
	// log.Warnf("destroy: %+v", u.DestroyConf)
	u.GetOutputsConf = nil
}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := utils.YAMLInterfaceToType(spec, u)
	if err != nil {
		return err
	}
	// Read manifests.
	baseDir := filepath.Join(config.Global.WorkingDir, u.StackPtr.TemplateDir)
	manifestsPath := filepath.Join(baseDir, u.Path)
	isDir, err := utils.CheckDir(manifestsPath)
	if isDir {
		if err != nil {
			return fmt.Errorf("read unit '%v': check path: %w", u.Name(), err)
		}
		u.ManifestsFiles.ReadDir(manifestsPath, baseDir, `.ya{0,1}ml$`)
		// log.Debugf("List %v", u.ManifestsFiles.SPrintLs())
		if u.ManifestsFiles.IsEmpty() {
			return fmt.Errorf("read unit '%v': no manifests found in path %v", u.Name(), u.Path)
		}
		u.recursive = true
	} else {
		err = u.ManifestsFiles.ReadFile(manifestsPath, baseDir)
		if err != nil {
			return fmt.Errorf("read unit '%v': read manifest: %w", u.Name(), err)
		}
	}

	for i, f := range *u.ManifestsFiles {
		if u.ApplyTemplate {
			templattedFile, errIsWarn, err := u.Stack().TemplateTry([]byte(f.Content))
			if err != nil {
				if !errIsWarn {
					return fmt.Errorf("read unit '%v': template manifest file: %w", f.FileName, err)
				}
			}
			(*(u).ManifestsFiles)[i].Content = string(templattedFile)
		}
	}
	u.UnitKind = u.KindKey()
	// u.StatePtr.ManifestsFiles = u.ManifestsFiles
	u.CacheDir = filepath.Join(u.Project().CodeCacheDir, u.Key())
	u.fillShellUnit()
	u.Prepare()
	return err
}

// Init unit.
func (u *Unit) Init() error {
	return nil
}

// Apply unit.
func (u *Unit) Apply() error {
	err := u.Unit.Apply()
	if err != nil {
		return err
	}
	u.CreateFiles = nil
	return nil
}

// Plan unit.
func (u *Unit) Plan() error {
	return nil
}

// Destroy unit.
func (u *Unit) Destroy() error {
	err := u.Unit.Destroy()
	if err != nil {
		return err
	}
	u.CreateFiles = nil
	return nil
}

func (u *Unit) GetManifestsMap() (res map[string]interface{}) {
	res = make(map[string]interface{})
	for _, file := range *u.ManifestsFiles {
		mns, err := utils.ReadYAMLObjects([]byte(file.Content))
		if err != nil {
			return nil
		}
		type manifestKey struct {
			Kind     string `yaml:"kind"`
			Metadata struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"metadata"`
		}
		for _, obj := range mns {
			keyS := manifestKey{}
			err = utils.YAMLInterfaceToType(obj, &keyS)
			if err != nil {
				return nil
			}
			if keyS.Metadata.Namespace == "" {
				keyS.Metadata.Namespace = "<default namespace>"
			}
			key := fmt.Sprintf("%s-->%s.%s.%s", file.FileName, keyS.Kind, keyS.Metadata.Namespace, keyS.Metadata.Name)
			res[key] = obj
		}
	}
	return
}

// ScanData scan all markers in unit, and build project unit links, and unit dependencies.
func (u *Unit) ScanData(scanner project.MarkerScanner) error {
	if err := u.Unit.ScanData(scanner); err != nil {
		return fmt.Errorf("prepare kubectl unit data: %w", err)
	}
	for _, file := range *u.ManifestsFiles {
		mns, err := utils.ReadYAMLObjects([]byte(file.Content))
		if err != nil {
			return err
		}
		var scannedFile []byte
		for i, mn := range mns {
			if i != 0 {
				scannedFile = append(scannedFile, []byte(fmt.Sprint("---\n"))...)
			}
			err := project.ScanMarkers(mn, scanner, u)
			if err != nil {
				return err
			}
			fmn, err := yaml.Marshal(mn)
			scannedFile = append(scannedFile, fmn...)
			// log.Debugf("Scanned file: %v", string(scannedFile))
		}
		file.Content = string(scannedFile)
	}
	return nil
}

// Prepare scan all markers in unit, and build project unit links, and unit dependencies.
func (u *Unit) Prepare() error {
	return u.ScanData(project.OutputsScanner)
}

func (u *Unit) Build() error {
	// Save state before output markers replace.
	u.SavedState = u.GetState()
	err := u.ScanData(project.OutputsReplacer)
	if err != nil {
		return err
	}
	// u.CreateFiles = u.ManifestsFiles
	err = u.Unit.Build()
	if err != nil {
		return err
	}
	return u.ManifestsFiles.WriteFiles(u.CacheDir)
}
