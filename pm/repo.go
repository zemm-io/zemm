package pm

import (
	"fmt"
	"net/url"
	"path"

	"encoding/json"

	"github.com/zemm-io/zemm/common"
)

type ListOrPackage struct {
	List    string `json:"list,omitempty",yaml:"list,omitempty"`
	Package string `json:"package,omitempty",yaml:"package,omitempty"`
}

type RInfo struct {
	Name        string          `json:"name",yaml:"name"`
	Version     string          `json:"version",yaml:"version"`
	Description string          `json:"description",yaml:"description"`
	Deprecation string          `json:"deprecation,omitempty",yaml:"deprecation,omitempty"`
	Depends     []ListOrPackage `json:"depends,omitempty",yaml:"depends,omitempty"`
	Supports    []ListOrPackage `json:"supports,omitempty",yaml:"supports,omitempty"`
}

type RPDependency struct {
	Package string `json:"package",yaml:"package"`
	Default string `json:"default,omitempty",yaml:"default,omitempty"`
}

type RPackage struct {
	Repository   *Repository    `json:"-",yaml:"-"`
	Name         string         `json:"name",yaml:"name"`
	Deprecation  string         `json:"deprecation",yaml:"deprecation"`
	Description  string         `json:"description",yaml:"description"`
	Author       string         `json:"author",yaml:"author"`
	Packager     string         `json:"packager,omitempty",yaml:"packager,omitempty"`
	License      string         `json:"license",yaml:"license"`
	Homepage     string         `json:"homepage",yaml:"homepage"`
	Repo         string         `json:"repo",yaml:"repo"`
	Provides     []string       `json:"provides",yaml:"provides"`
	Dependencies []RPDependency `json:"dependencies",yaml:"dependencies"`
}

type Repository struct {
	index        string     `json:"-",yaml:"-"`
	list         string     `json:"-",yaml:"-"`
	isDependency bool       `json:"-",yaml:"-"`
	Info         RInfo      `json:"info",yaml:"info"`
	Packages     []RPackage `json:"packages",yaml:"packages"`
}

func NewRepository(index, list string) (*Repository, error) {
	r := &Repository{index: index, list: list, isDependency: false}
	err := r.Update()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) Update() error {
	iu, err := r.GetFullURL()
	if err != nil {
		return err
	}

	err = common.URLToStruct(iu, &r)
	if err != nil {
		return err
	}

	// Link packages to this repository to locate the repo from the package
	for i := range r.Packages {
		r.Packages[i].Repository = r
	}

	return nil
}

func (r *Repository) GetIndex() string {
	return r.index
}

func (r *Repository) GetList() string {
	return r.list
}

func (r *Repository) GetFullURL() (string, error) {
	if common.URLIsValidAndHTTP(r.index) {
		// Is http/s URL
		u, _ := url.Parse(r.index) // can ignore err here cause of "URLIsValidAndHTTP"
		u.Path = path.Join(u.Path, "lists", r.list)
		return u.String(), nil
	}

	j := path.Join(r.index, "lists", r.list)
	if common.FileExists(j) {
		return j, nil
	}
	if common.FileExists(j + ".yaml") {
		return j + ".yaml", nil
	}
	if common.FileExists(j + ".json") {
		return j + ".json", nil
	}

	return "", fmt.Errorf("File %v doesn't exists", j)
}

func (r *Repository) IsDependency() bool {
	return r.isDependency
}

func (r *Repository) SetDependency(isDep bool) {
	r.isDependency = isDep
}

func (r *Repository) DumpJSON() ([]byte, error) {
	return json.Marshal(r)
}
