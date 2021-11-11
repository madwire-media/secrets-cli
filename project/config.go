package project

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/madwire-media/secrets-cli/vars"
	"github.com/ryanuber/go-glob"
	"gopkg.in/yaml.v3"
)

// ClassUpdate represents a CLI change in locally-defined secret classes
type ClassUpdate struct {
	Reset bool
	FilterOptions
}

// Project represents an environment with a secrets.yaml, similar to a Git
// repository
type Project struct {
	Config
	classes      FilterOptions
	path         string
	lastState    LockState
	currentState LockState
}

// Config is the root configuration for a secrets.yaml file
type Config struct {
	Secrets []SecretConfig `yaml:"secrets"`
}

// OpenProject opens a project based on the current working directory. It will
// traverse the current directory and every parent directory until it finds a
// secrets.yaml file. It also loads the class file and lockfile after a project
// has been found
func OpenProject() (*Project, error) {
	project := Project{
		path: vars.Workdir,
	}

	for {
		_, err := os.Stat(filepath.Join(project.path, "secrets.yaml"))

		if err == nil {
			break
		} else if !os.IsNotExist(err) {
			return nil, err
		}

		parent := filepath.Dir(project.path)

		if parent == project.path {
			return nil, errors.New("could not find a secrets manifest in working directory or parent directories")
		}

		project.path = parent
	}

	text, err := ioutil.ReadFile(filepath.Join(project.path, "secrets.yaml"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(text, &project.Config)
	if err != nil {
		return nil, err
	}

	secretFilenames := make(map[string]struct{})

	for _, secret := range project.Config.Secrets {
		if _, ok := secretFilenames[secret.File]; ok {
			return nil, fmt.Errorf("duplicate filename in config: %s", secret.File)
		}

		secretFilenames[secret.File] = struct{}{}
	}

	if err := project.loadClasses(); err != nil {
		return nil, err
	}

	if err := project.readLockfile(); err != nil {
		return nil, err
	}

	return &project, nil
}

func (project *Project) Save() error {
	text, err := yaml.Marshal(&project.Config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(project.path, "secrets.yaml"),
		text,
		0666,
	)
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) applyClassUpdate(update ClassUpdate) error {
	if update.Reset {
		project.classes = FilterOptions{}
	}

	if update.DefaultAll {
		project.classes = FilterOptions{
			DefaultAll: true,
		}
	}

	for _, class := range update.Add {
		var removed bool

		project.classes.Subtract, removed = removeString(project.classes.Subtract, class)

		if !project.classes.DefaultAll && !removed {
			if !containsString(project.classes.Add, class) {
				project.classes.Add = append(project.classes.Add, class)
			}
		}
	}

	for _, class := range update.Subtract {
		var removed bool

		project.classes.Add, removed = removeString(project.classes.Add, class)

		if project.classes.DefaultAll && !removed {
			if !containsString(project.classes.Subtract, class) {
				project.classes.Subtract = append(project.classes.Subtract, class)
			}
		}
	}

	if !vars.IsCICD {
		err := project.saveClasses()
		if err != nil {
			return err
		}
	}

	return nil
}

func (project *Project) saveClasses() error {
	// Don't worry about saving the .secretclasses file in CI/CD mode
	if vars.IsCICD {
		return nil
	}

	classString := ""

	if project.classes.DefaultAll {
		classString += "+all"

		for _, subtracted := range project.classes.Subtract {
			classString += ",-" + subtracted
		}
	} else {
		first := true

		for _, added := range project.classes.Add {
			if first {
				first = false
			} else {
				classString += ","
			}

			classString += "+" + added
		}
	}

	shouldWrite := true
	filename := filepath.Join(project.path, ".localsecretclasses")

	if classString == "" {
		if _, err := os.Stat(filename); err != nil {
			shouldWrite = false
		}
	}

	if shouldWrite {
		err := project.ensureClassfileInGitignore()
		if err != nil {
			return err
		}

		classString += "\n"

		err = ioutil.WriteFile(filename, []byte(classString), 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (project *Project) loadClasses() error {
	classes := FilterOptions{
		Add:      []string{},
		Subtract: []string{},
	}

	if !vars.IsCICD {
		filename := filepath.Join(project.path, ".localsecretclasses")

		classBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil
		}

		for _, class := range strings.Split(string(classBytes), ",") {
			class = strings.TrimSpace(class)

			if class == "+all" {
				classes.DefaultAll = true
			} else if strings.HasPrefix(class, "+") {
				classes.Add = append(classes.Add, class[1:])
			} else if strings.HasPrefix(class, "-") {
				classes.Subtract = append(classes.Subtract, class[1:])
			}
		}
	}

	project.classes = classes

	return nil
}

func (project *Project) ensureClassfileInGitignore() error {
	// Don't worry about changing the .gitignore in CI/CD mode
	if vars.IsCICD {
		return nil
	}

	filename := filepath.Join(project.path, ".gitignore")

	gitignore, err := ioutil.ReadFile(filename)
	var newGitignore string

	if os.IsNotExist(err) {
		newGitignore = "/.localsecretclasses\n"
	} else {
		scanner := bufio.NewScanner(bytes.NewReader(gitignore))
		for scanner.Scan() {
			text := scanner.Text()

			hashIdx := strings.Index(text, "#")
			if hashIdx > 0 {
				text = text[hashIdx:]
			}

			text = strings.TrimSpace(text)
			text = strings.TrimLeft(text, "/")

			// .gitignore already has matching line for .localsecretclasses
			if glob.Glob(text, ".localsecretclasses") {
				return nil
			}
		}

		newGitignore = string(gitignore)

		if newGitignore[len(newGitignore)-1:] != "\n" {
			newGitignore += "\n"
		}

		newGitignore += "/.localsecretclasses\n"
	}

	return ioutil.WriteFile(filename, []byte(newGitignore), 0664)
}

func removeString(slice []string, str string) ([]string, bool) {
	removed := false

	for i := 0; i < len(slice); i++ {
		if slice[i] == str {
			slice[i] = slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			removed = true
			i--
		}
	}

	return slice, removed
}

func containsString(slice []string, str string) bool {
	for _, sliceStr := range slice {
		if sliceStr == str {
			return true
		}
	}

	return false
}
