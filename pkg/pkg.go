package pkg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/zemm-io/zemm/common"

	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
)

type InfoDep struct {
	Package string `json:"package" yaml:"package"`
}

type Info struct {
	Name         string    `json:"name" yaml:"name"`
	Version      string    `json:"version" yaml:"version"`
	Description  string    `json:"description" yaml:"description"`
	Author       string    `json:"author" yaml:"author"`
	Packager     string    `json:"packager,omitempty" yaml:"packager,omitempty"`
	License      string    `json:"license" yaml:"license"`
	Homepage     string    `json:"homepage" yaml:"homepage"`
	Repo         string    `json:"repo" yaml:"repo"`
	Provides     []string  `json:"provides" yaml:"provides"`
	Dependencies []InfoDep `json:"dependencies" yaml:"dependencies"`
	Recommends   []InfoDep `json:"recommends" yaml:"recommends"`
}

type FileOrDir struct {
	File      string `json:"file,omitempty" yaml:"file,omitempty"`
	Directory string `json:"dir,omitempty" yaml:"dir,omitempty"`
}

type Pkg struct {
	path  string
	Info  Info        `json:"info" yaml:"info"`
	Files []FileOrDir `json:"files" yaml:"files"`
}

func NewPkg(path string) (*Pkg, error) {
	p := &Pkg{path: path}
	err := p.Parse()
	return p, err
}

func (p *Pkg) Parse() error {
	if common.URLIsValidAndHTTP(p.path) {
		return fmt.Errorf("Loading packages from URL is not supported (yet)")
	}

	if !common.FileExists(p.path) {
		if common.FileExists(p.path + ".yaml") {
			p.path = p.path + ".yaml"
		} else if common.FileExists(p.path + ".json") {
			p.path = p.path + ".json"
		} else {
			return fmt.Errorf("File \"%s\", does not exists", p.path)
		}
	}

	return common.URLToStruct(p.path, &p)
}

func (p *Pkg) Namespace() string {
	return strings.Split(p.Info.Name, "/")[0]
}

func (p *Pkg) Package() string {
	return strings.Split(p.Info.Name, "/")[1]
}

func (p *Pkg) verifyFiles(result *multierror.Error) {
	basePath := path.Dir(p.path)

	for _, fod := range p.Files {
		if fod.File != "" {
			if !common.FileExists(path.Join(basePath, fod.File)) {
				result = multierror.Append(result, fmt.Errorf("File \"%s\" doesn't exists", path.Join(basePath, fod.File)))
			}
		} else if fod.Directory != "" {
			if !common.DirExists(path.Join(basePath, fod.Directory)) {
				result = multierror.Append(result, fmt.Errorf("Directory \"%s\" doesn't exists", path.Join(basePath, fod.Directory)))
			}
		}
	}

	return
}

// Verify verifies the package
func (p *Pkg) Verify() error {
	result := &multierror.Error{}

	if !strings.Contains(p.Info.Name, "/") {
		result = multierror.Append(result, errors.New("Name must contain a namespace, for example \"mynamespace/mypackage\""))
	}

	p.verifyFiles(result)

	return result.ErrorOrNil()
}

func (p *Pkg) MakePackage() (string, error) {
	basePath, err := filepath.Abs(path.Dir(p.path))
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "zemmpkg-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	// Copy zemmpkg.yaml
	if err = common.CopyFile(path.Join(basePath, "zemmpkg.yaml"), path.Join(tmpDir, "zemmpkg.yaml"), true); err != nil {
		return "", err
	}

	// Copy files and directories
	paths := []string{"zemmpkg.yaml"}
	for _, fod := range p.Files {
		if fod.File != "" {
			// Copy the file to the tmp path
			if err = copy.Copy(path.Join(basePath, fod.File), path.Join(tmpDir, fod.File)); err != nil {
				return "", err
			}

			// Add each directory
			if strings.Contains(fod.File, "/") {
				exp := strings.Split(path.Dir(fod.File), "/")
				for i := range exp {
					if i == 0 {
						paths = append(paths, exp[0])
					} else {
						paths = append(paths, path.Join(exp[0:i]...))
					}
				}
			}
			paths = append(paths, fod.File)
		} else if fod.Directory != "" {
			// Copy the directory to the tmp path
			if err = copy.Copy(path.Join(basePath, fod.Directory), path.Join(tmpDir, fod.Directory)); err != nil {
				return "", err
			}

			// Add each directory
			if strings.Contains(fod.Directory, "/") {
				exp := strings.Split(fod.Directory, "/")
				for i := range exp {
					if i > 0 && len(exp[i]) > 0 {
						paths = append(paths, path.Join(exp[0:i]...))
					}
				}
			}
			paths = append(paths, fod.Directory)
		}
	}

	// Create archive name
	outDir, err := ioutil.TempDir("", "zemmout-")
	if err != nil {
		return "", err
	}
	// defer os.RemoveAll(outDir)

	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err = os.Chdir(tmpDir); err != nil {
		return "", err
	}
	archFilePath := path.Join(outDir, fmt.Sprintf("%s-%s.txz", p.Package(), p.Info.Version))

	if err = archiver.Archive([]string{"."}, archFilePath); err != nil {
		return "", err
	}

	if err = os.Chdir(pwd); err != nil {
		return "", err
	}

	return archFilePath, nil
}
