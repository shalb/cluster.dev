package ui

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/apex/log"
	"github.com/paulrademacher/climenu"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Generator struct {
	renderedFiles        map[string][]byte
	selectedTemplateName string
	categoryDir          string
	templateDir          string
	templateDataDir      string
	dataForTmpl          map[string]interface{}
	templateConfig       templateConfSpec
	categoryConfig       categoryConfSpec
	interactive          bool
	templateFs           TmplFS
}

//go:embed templates/*
var embadedTemplatesFs embed.FS

const templatesDir = "templates"

type optsSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Regex       string `yaml:"regex,omitempty"`
	Default     string `yaml:"default,omitempty"`
}

type categoryConfSpec struct {
	Header   string `yaml:"header"`
	Question string `yaml:"question"`
	Default  string `yaml:"default"`
}

type templateConfSpec struct {
	Description string     `yaml:"description"`
	Options     []optsSpec `yaml:"options"`
	Replacers   []replacer `yaml:"filenames_replace"`
	HelpMessage string     `yaml:"help_message,omitempty"`
}
type replacer struct {
	Regex   string `yaml:"regex"`
	VarName string `yaml:"replace_var_name"`
}

func CreateSecret() error {
	generator, err := NewGeneratorLocal("secret")
	if err != nil {
		return fmt.Errorf("new secret: %v", err.Error())
	}
	generator.SetInteractive()
	escaped, err := generator.RunMainMenu()
	if err != nil {
		return fmt.Errorf("new secret: %v", err.Error())
	}
	if escaped {
		log.Warn("Exiting...")
		os.Exit(0)
	}
	secretDriver, exists := project.SecretDriversMap[generator.SelectedTemplate()]
	if !exists {
		return fmt.Errorf("new secret: secret driver '%v' does not exists ", generator.SelectedTemplate())
	}
	err = generator.RunTemplate()
	if err != nil {
		return fmt.Errorf("new secret: %v", err.Error())
	}
	err = secretDriver.Create(generator.RenderedFiles())
	if err != nil {
		return fmt.Errorf("new secret: %v", err.Error())
	}
	return nil
}

// GetProjectTemplates returns list of templates in dir.
func GetProjectTemplates(gitURL string) (res []string, err error) {
	generator, err := NewGeneratorRemote(gitURL)
	if err != nil {
		return nil, fmt.Errorf("new project: %v", err.Error())
	}
	return getDirSubCats(generator.categoryDir, generator.templateFs)
}

func CreateProject(dir, gitURL string, args ...string) error {
	generator, err := NewGeneratorRemote(gitURL)
	if err != nil {
		return fmt.Errorf("new project: %v", err.Error())
	}
	if config.Global.Interactive {
		generator.SetInteractive()
	}
	escaped, err := generator.RunMainMenu(args...)
	if err != nil {
		return fmt.Errorf("new project: %v", err.Error())
	}
	if escaped {
		log.Warn("Exiting...")
		os.Exit(0)
	}
	err = generator.RunTemplate()
	if err != nil {
		return fmt.Errorf("new project: %v", err.Error())
	}
	err = generator.WriteFiles(dir)
	if err != nil {
		return fmt.Errorf("new project: %v", err.Error())
	}
	return nil
}

func (g *Generator) RenderedFiles() map[string][]byte {
	return g.renderedFiles
}

func (g *Generator) SelectedTemplate() string {
	return g.selectedTemplateName
}

func NewGeneratorLocal(categoryName string) (g *Generator, err error) {
	catDir := filepath.Join(templatesDir, categoryName)
	return newGenerator(catDir, embadedTemplatesFs)
}

func NewGeneratorRemote(gitSrc string) (g *Generator, err error) {
	catDir, err := GetTemplateGenerators(gitSrc)
	if err != nil {
		return
	}
	return newGenerator(".", NewTmplFS(catDir))
}

func newGenerator(catDir string, tFS TmplFS) (g *Generator, err error) {
	categoryConf, err := getCategorySpec(catDir, tFS)
	if err != nil {
		return nil, err
	}
	g = &Generator{
		categoryDir:    catDir,
		categoryConfig: categoryConf,
		renderedFiles:  make(map[string][]byte),
		dataForTmpl:    make(map[string]interface{}),
		interactive:    false,
		templateFs:     tFS,
	}
	return g, nil
}

func (g *Generator) SetInteractive() {
	g.interactive = true
}

func (g *Generator) RunMainMenu(subCategory ...string) (escaped bool, err error) {
	categoryTmplList, err := getDirSubCats(g.categoryDir, g.templateFs)
	if err != nil {
		return
	}
	log.Debugf("RunMainMenu :%v", categoryTmplList)
	menu := climenu.NewButtonMenu(g.categoryConfig.Header, g.categoryConfig.Question)
	generatorSpecs := map[string]templateConfSpec{}
	for _, tmplName := range categoryTmplList {
		sp, err := getTemplateSpec(g.categoryDir, tmplName, g.templateFs)
		if err != nil {
			log.Fatal(err.Error())
		}
		menu.AddMenuItem(sp.Description, tmplName)
		generatorSpecs[tmplName] = sp
	}
	var keysList string
	i := 0
	for k := range generatorSpecs {
		if i != 0 {
			keysList += "\n"
		}
		keysList += k
		i++
	}
	if g.interactive {
		g.selectedTemplateName, escaped = menu.Run()
		if escaped {
			return
		}
	} else {
		if len(subCategory) != 1 {
			if g.categoryConfig.Default == "" {
				return false, fmt.Errorf("generator: unexpected category param %v, expected 1 string, use one of: \n%v", len(subCategory), keysList)
			}
			g.selectedTemplateName = g.categoryConfig.Default
		} else {
			g.selectedTemplateName = subCategory[0]
		}
	}

	var exists bool
	g.templateConfig, exists = generatorSpecs[g.selectedTemplateName]
	if !exists {
		return false, fmt.Errorf("generator: template '%v' does not found, use one of: \n%v", g.selectedTemplateName, keysList)
	}
	g.templateDir = filepath.Join(g.categoryDir, g.selectedTemplateName)
	g.templateDataDir = filepath.Join(g.templateDir, "data")
	if g.templateConfig.HelpMessage != "" && g.interactive {
		ClearScreen()
		fmt.Println(g.templateConfig.HelpMessage)
		respond := climenu.GetText("Continue?(yes/no)", "yes")
		if respond == "yes" {
			escaped = false
		} else {
			escaped = true
		}
	}
	return
}

func (g *Generator) RunTemplate() (err error) {
	for _, opt := range g.templateConfig.Options {
		if opt.Regex == "" {
			opt.Regex = ".*"
		}
		if !g.interactive {
			g.dataForTmpl[opt.Name] = opt.Default
			continue
		}
		for {
			respond := climenu.GetText(opt.Description, "")
			if verifyRegex(respond, opt.Regex) {
				g.dataForTmpl[opt.Name] = respond
				break
			}
			log.Warnf("Data verifycation error, regex fo check: '%v'", opt.Regex)
		}
	}
	err = g.compileTree(g.templateDataDir)
	return
}

func (g *Generator) WriteFiles(path string) (err error) {
	for outputFileName, fileData := range g.renderedFiles {
		filename := filepath.Join(path, outputFileName)
		fileDir := filepath.Join(path, filepath.Dir(outputFileName))
		err = os.MkdirAll(fileDir, os.ModePerm)
		if err != nil {
			return
		}
		os.WriteFile(filename, fileData, fs.ModePerm)
		if err != nil {
			return
		}
		log.Infof("Creating: %v", filepath.Base(outputFileName))
	}
	return
}

func (g *Generator) compileTree(path string, relPath ...string) (err error) {
	dir, err := g.templateFs.ReadDir(path)
	if err != nil {
		return
	}
	for _, elem := range dir {
		if elem.IsDir() {
			err = g.compileTree(filepath.Join(path, elem.Name()), append(relPath, elem.Name())...)
			if err != nil {
				return
			}
			continue
		}
		inputFileName := filepath.Join(path, elem.Name())
		outputFileName := filepath.Join(filepath.Join(relPath...), replaceFilename(elem.Name(), g.templateConfig.Replacers, g.dataForTmpl))
		var tmplFileRaw []byte
		tmplFileRaw, err = g.templateFs.ReadFile(inputFileName)
		if err != nil {
			err = fmt.Errorf("internal error, %v", err.Error())
			return
		}
		g.renderedFiles[outputFileName], err = g.applyTemplate(tmplFileRaw)
		if err != nil {
			return
		}
	}
	return
}

func (g *Generator) applyTemplate(tmplRaw []byte) ([]byte, error) {
	tmpl, err := template.New("main").Delims("/{", "}/").Option("missingkey=error").Parse(string(tmplRaw))
	if err != nil {
		return nil, fmt.Errorf("render template file: %v", err.Error())
	}
	result := bytes.Buffer{}
	err = tmpl.Execute(&result, g.dataForTmpl)
	if err != nil {
		return nil, fmt.Errorf("render template file: %v", err.Error())
	}
	return result.Bytes(), nil
}

func (g *Generator) applyTemplateString(tmplRaw string) (string, error) {
	tmpl, err := template.New("main").Delims("/{", "}/").Option("missingkey=error").Parse(tmplRaw)
	if err != nil {
		return "", fmt.Errorf("render template file: %v", err.Error())
	}
	result := bytes.Buffer{}
	err = tmpl.Execute(&result, g.dataForTmpl)
	if err != nil {
		return "", fmt.Errorf("render template file: %v", err.Error())
	}
	return result.String(), nil
}

// Read directories names in path. If non dir founded - return error.
func getDirSubCats(path string, tFS TmplFS) (cats []string, err error) {
	dir, err := tFS.ReadDir(path)
	if err != nil {
		err = fmt.Errorf("reading templates: internal error: %v", err.Error())
		return
	}
	for _, elem := range dir {
		if !elem.IsDir() {
			continue
		}
		cats = append(cats, elem.Name())
	}
	return
}

func getTemplateSpec(catDir, templateName string, tFS TmplFS) (res templateConfSpec, err error) {
	rs := templateConfSpec{}
	confFileName := filepath.Join(catDir, templateName, "config.yaml")
	// log.Warn(confFileName)
	configRaw, err := tFS.ReadFile(confFileName)
	if err != nil {
		err = fmt.Errorf("reading template: internal error (file not found %v): %v", confFileName, err.Error())
		return
	}
	err = yaml.Unmarshal(configRaw, &rs)
	if err != nil {
		err = fmt.Errorf("reading template: config parse '%v': %v", confFileName, utils.ResolveYamlError(configRaw, err))
		return
	}
	res = rs
	return
}

func getCategorySpec(catDir string, tFS TmplFS) (res categoryConfSpec, err error) {
	rs := categoryConfSpec{}
	confFileName := filepath.Join(catDir, "config.yaml")
	configRaw, err := tFS.ReadFile(confFileName)
	if err != nil {
		err = fmt.Errorf("reading template: file not found %v: %v", confFileName, err.Error())
		return
	}
	err = yaml.Unmarshal(configRaw, &rs)
	if err != nil {
		err = fmt.Errorf("reading template: config parse %v: %v", confFileName, utils.ResolveYamlError(configRaw, err))
		return
	}
	res = rs
	return
}

func verifyRegex(name, regex string) bool {
	return regexp.MustCompile(regex).MatchString(name)
}

func replaceFilename(fn string, rps []replacer, vars map[string]interface{}) string {
	fileName := filepath.Base(fn)
	path := filepath.Dir(fn)
	for _, rp := range rps {
		log.Debugf("File: %v; Replacer: %v; vars: %v", fileName, rp, vars)
		replTo, exists := vars[rp.VarName]
		if !exists {
			return fn
		}
		res := regexp.MustCompile(rp.Regex).ReplaceAllString(fileName, replTo.(string))
		log.Debugf("Result: %v", res)
		if res != fileName {
			return filepath.Join(path, res)
		}
	}
	return fn
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func GetTemplateGenerators(tmplSrc string) (tmplDir string, err error) {
	if !config.Global.UseCache {
		os.RemoveAll(config.Global.TemplatesCacheDir)
	}
	err = os.MkdirAll(config.Global.TemplatesCacheDir, os.ModePerm)
	if err != nil {
		return
	}
	dr, err := utils.GetTemplate(tmplSrc, config.Global.TemplatesCacheDir, "generator")
	if err != nil {
		err = fmt.Errorf("download template: %v\n   See details about stack template reference: https://docs.cluster.dev/structure-stack/", err.Error())
		return
	}
	log.Debugf("Generator: template dir: %v", dr)

	tmplDir = filepath.Join(dr, ".cdev-metadata/generator")
	return
}
