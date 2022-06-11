package base

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/common"
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

// Unit describe cluster.dev unit to deploy/destroy k8s resources with kubectl.
type Unit struct {
	common.Unit
	Namespace          string             `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Kubeconfig         string             `yaml:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	KubectlOpts        string             `yaml:"kubectl_opts,omitempty" json:"kubectl_opts,omitempty"`
	KubectlCliConf     *KubectlCliT       `yaml:"kubectl,omitempty" json:"kubectl,omitempty"`
	Path               string             `yaml:"path" json:"path"`
	ManifestsFiles     *common.FilesListT `yaml:"-" json:"manifests"`
	ApplyTemplate      bool               `yaml:"apply_template" json:"-"`
	recursive          bool               `yaml:"-" json:"-"`
	UnitKind           string             `yaml:"-" json:"type"`
	SavedState         interface{}        `yaml:"-" json:"-"`
	manifestsForDelete *common.FilesListT `yaml:"-" json:"-"`
	CreateNamespaces   bool               `yaml:"create_namespaces" json:"-"`
	createNSList       []string           `yaml:"-" json:"-"`
}

var kubectlBin = "kubectl"

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) fillShellUnit() {
	commandOpts := "-R "
	// if u.recursive {
	// 	commandOpts += "-R "
	// }
	if u.Namespace != "" {
		commandOpts += "-n " + u.Namespace
	}
	if u.KubectlOpts != "" {
		commandOpts = fmt.Sprintf("%s %s", commandOpts, u.KubectlOpts)
	}
	if u.Kubeconfig != "" {
		commandOpts = fmt.Sprintf("%s --kubeconfig='%s'", commandOpts, u.Kubeconfig)
	}
	if u.manifestsForDelete != nil {
		u.ApplyConf = &common.OperationConfig{
			Commands: []interface{}{
				fmt.Sprintf("%s delete %s -f %s", kubectlBin, commandOpts, filepath.Join(u.CacheDir, "delete/main.yaml")),
			},
		}
	} else {
		u.ApplyConf = &common.OperationConfig{}
	}
	u.ApplyConf.Commands = append(u.ApplyConf.Commands, fmt.Sprintf("%s apply %s -f %s", kubectlBin, commandOpts, filepath.Join(u.CacheDir, "workdir")))

	// log.Warnf("path: %v", u.Path)
	u.DestroyConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s delete %s -R -f %s", kubectlBin, commandOpts, filepath.Join(u.CacheDir, "workdir")),
		},
	}
	// u.CreateFiles = nil
	// log.Warnf("apply: %+v", u.ApplyConf)
	// log.Warnf("destroy: %+v", u.DestroyConf)
	u.GetOutputsConf = nil
}

func (u *Unit) createNamespacesIfNotExists() error {
	rn, err := executor.NewExecutor(u.CacheDir)
	if err != nil {
		log.Debug(err.Error())
		return fmt.Errorf("create namespace: %w", err)
	}
	if len(u.createNSList) > 0 {
		for _, ns := range u.createNSList {
			kubeconfigOpt := ""
			if u.Kubeconfig != "" {
				kubeconfigOpt = fmt.Sprintf("--kubeconfig='%s'", u.Kubeconfig)
			}
			cmd := fmt.Sprintf("%s %s create ns %s", kubectlBin, kubeconfigOpt, ns)
			_, errMsg, err := rn.Run(cmd)
			if len(errMsg) > 1 {
				if err != nil {
					log.Debugf("Failed attempt to create namespace (ignore) %v", string(errMsg))
				}
			}
		}
	}
	return nil
}

func (u *Unit) ReadManifestsPath(src string) error {
	// Check if path is URL or local dir.
	var manifestsPath string
	baseDir := filepath.Join(config.Global.WorkingDir, u.StackPtr.TemplateDir)
	if utils.IsLocalPath(src) {
		if utils.IsAbsolutePath(src) {
			manifestsPath = src
		} else {
			manifestsPath = filepath.Join(baseDir, src)
		}
		isDir, err := utils.CheckDir(manifestsPath)
		if err != nil {
			return fmt.Errorf("check path: %w", err)
		}
		if isDir {
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
		log.Debugf("Template dir: %v", manifestsPath)

	} else {
		manifest, err := utils.GetFileByUrl(src)
		if err != nil {
			return fmt.Errorf("get remote file: %w", err)
		}
		err = u.ManifestsFiles.Add("./main.yaml", manifest, fs.ModePerm)
		if err != nil {
			return fmt.Errorf("add remote file: %w", err)
		}
	}
	return nil

}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := utils.YAMLInterfaceToType(spec, u)
	if err != nil {
		return err
	}
	err = u.ReadManifestsPath(u.Path)
	if err != nil {
		return fmt.Errorf("read unit '%v': read manifests: %w", u.Name(), err)
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
	return err
}

// Init unit.
func (u *Unit) Init() error {
	return nil
}

// Apply unit.
func (u *Unit) Apply() error {
	err := u.createNamespacesIfNotExists()
	if err != nil {
		return err
	}
	err = u.Unit.Apply()
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

func (u *Unit) GetManifestsMap() (res map[string]interface{}, namespaces []string) {
	res = make(map[string]interface{})
	namespaces = []string{}
	nsUniq := map[string]bool{}
	for _, file := range *u.ManifestsFiles {
		mns, err := utils.ReadYAMLObjects([]byte(file.Content))
		if err != nil {
			return nil, nil
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
				return nil, nil
			}
			if keyS.Metadata.Namespace != "" {
				if exists := nsUniq[keyS.Metadata.Namespace]; !exists {
					nsUniq[keyS.Metadata.Namespace] = true
					namespaces = append(namespaces, keyS.Metadata.Namespace)
				}
			}
			if keyS.Metadata.Namespace == "" {
				keyS.Metadata.Namespace = "<default namespace>"
			}
			key := fmt.Sprintf("%s.%s.%s", keyS.Kind, keyS.Metadata.Namespace, keyS.Metadata.Name)
			res[key] = obj
		}
	}
	if u.Namespace != "" {
		if exists := nsUniq[u.Namespace]; !exists {
			namespaces = append(namespaces, u.Namespace)
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
	err := project.ScanMarkers(&u.Kubeconfig, scanner, u)
	if err != nil {
		return err
	}
	return nil
}

// Prepare scan all markers in unit, and build project unit links, and unit dependencies.
func (u *Unit) Prepare() error {
	err := u.Unit.Prepare()
	if err != nil {
		return err
	}
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
	if u.ProjectPtr.OwnState != nil {
		stateUnit, exists := u.ProjectPtr.OwnState.Units[u.Key()]
		deleteYAML := ""
		if exists {
			stateUnit, ok := stateUnit.(*Unit)
			if !ok {
				log.Fatalf("internal error")
			}
			stateManifests, _ := stateUnit.GetManifestsMap()
			manifests, createNSList := u.GetManifestsMap()
			u.createNSList = createNSList
			i := 0
			for key, man := range stateManifests {
				// log.Warnf("stateManifests: %v", key)
				if i > 0 {
					deleteYAML = deleteYAML + "\n---\n"
				}
				if _, ok := manifests[key]; !ok {
					manifestData, err := yaml.Marshal(man)
					if err != nil {
						return err
					}
					deleteYAML = deleteYAML + string(manifestData)
				}
			}
		}
		if deleteYAML != "" {
			u.manifestsForDelete = &common.FilesListT{}
			u.manifestsForDelete.Add("./delete/main.yaml", deleteYAML, fs.ModePerm)
			err := u.manifestsForDelete.WriteFiles(u.CacheDir)
			if err != nil {
				return err
			}
		}
	}

	u.fillShellUnit()
	return u.ManifestsFiles.WriteFiles(filepath.Join(u.CacheDir, "workdir"))
}
