package pm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/tpazderka/warning"
)

func filterPrintWarning(err error) error {
	if err == nil {
		return nil
	}

	result := &multierror.Error{}

	switch v := err.(type) {
	case warning.Warning:
		fmt.Printf("WARN: %v\n", err)
	case *multierror.Error:
		for _, merr := range v.WrappedErrors() {
			if warning.IsWarning(merr) {
				fmt.Printf("WARN: %v\n", merr)
			} else {
				result = multierror.Append(result, merr)
			}
		}
	default:
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func stringSliceContains(h []string, n string) bool {
	for _, k := range h {
		if k == n {
			return true
		}
	}

	return false
}

func TestDefaultPMAddRepository(t *testing.T) {
	pm, err := NewPackageManager()
	if err != nil {
		t.Error(err)
	}

	err = pm.AddRepository("../examples/repo/", "tuatzemm/suite/1.0.1")
	if err != nil {
		t.Error(err)
	}

	err = pm.Validate()
	if err != nil {
		t.Error(err)
	}

	repos := pm.GetRepositories()
	if len(repos) != 2 {
		t.Error(fmt.Errorf("Invalid number of repos: %d", len(repos)))
		return
	}
	if repos[0].GetList() != "tuatzemm/suite/1.0.0" {
		t.Error(fmt.Errorf("Invalid first repo: %s", repos[0].GetList()))
	}
	if repos[1].GetList() != "tuatzemm/suite/1.0.1" {
		t.Error(fmt.Errorf("Invalid second repo: %s", repos[1].GetList()))
	}
}

func TestDefaultPMAddMinadminRepository(t *testing.T) {
	pm, err := NewPackageManager()
	if err != nil {
		t.Error(err)
	}

	err = pm.AddRepository("../examples/repo/", "minadmin/minadmin/1.0.0")
	if err != nil {
		t.Error(err)
	}

	err = pm.Validate()
	if err != nil {
		t.Error(err)
	}

	repos := pm.GetRepositories()
	keys := make([]string, len(repos))
	for i, r := range repos {
		keys[i] = r.GetList()
	}
	if len(repos) != 5 {
		t.Error(fmt.Errorf("Invalid number of repos: %d, repos: %v", len(repos), keys))
		return
	}
	if repos[0].GetList() != "library/nats/2.1.9" {
		t.Error(fmt.Errorf("Invalid first repo: %s, repos: %v", repos[0].GetList(), keys))
	}
	if repos[1].GetList() != "library/postgres/13.2" {
		t.Error(fmt.Errorf("Invalid first repo: %s, repos: %v", repos[1].GetList(), keys))
	}
	if repos[2].GetList() != "tuatzemm/suite/1.0.0" {
		t.Error(fmt.Errorf("Invalid first repo: %s, repos: %v", repos[2].GetList(), keys))
	}
	if repos[3].GetList() != "tuatzemm/suite/1.0.1" {
		t.Error(fmt.Errorf("Invalid second repo: %s, repos %v", repos[3].GetList(), keys))
	}
	if repos[4].GetList() != "minadmin/minadmin/1.0.0" {
		t.Error(fmt.Errorf("Invalid third repo: %s, repos: %v", repos[4].GetList(), keys))
	}
}

func TestDependenciesOfMinadmin(t *testing.T) {
	pm, err := NewPackageManager()
	if err != nil {
		t.Error(err)
	}

	err = pm.AddRepository("../examples/repo/", "minadmin/minadmin/1.0.0")
	if err != nil {
		t.Error(err)
	}

	err = pm.Validate()
	if err != nil {
		t.Error(err)
	}

	pkgs, err := pm.GetDependencies([]string{"minadmin/minadmin_pgsql"}, true)
	err = filterPrintWarning(err)
	if err != nil {
		t.Error(err)
	}

	names := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		names[i] = pkg.Name
	}

	expectedResult := []string{
		"minadmin/minadmin_pgsql",
		"tuatzemm/auth_sql_pgsql",
		"tuatzemm/abac_pgsql",
		"library/nats",
		"library/postgres",
		"tuatzemm/auth",
		"tuatzemm/sql_pgsql",
	}

	if len(names) != len(expectedResult) {
		t.Error(fmt.Errorf("Got %d packages, excpected %d, packages: %v", len(names), len(expectedResult), names))
	}

	for _, ep := range expectedResult {
		if ok := stringSliceContains(names, ep); !ok {
			t.Error(fmt.Errorf("Package %s missing in result", ep))
		}
	}

}
