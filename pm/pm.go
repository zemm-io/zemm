package pm

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/tpazderka/warning"
)

type PackageManager struct {
	repos     []*Repository
	packages  map[string]*RPackage
	providers map[string]map[string]*RPackage
}

func (pm *PackageManager) addRepositoryWithExtends(index, list string, repos []*Repository, resultErr *multierror.Error) ([]*Repository, *multierror.Error) {
	r, err := NewRepository(index, list)
	if err != nil {
		return []*Repository{}, multierror.Append(resultErr, err)
	}
	repos = append(repos, r)

	if len(r.Info.Depends) == 0 {
		// Just add the repo and return
		return repos, resultErr
	}

	// Load all dependencies by calling ourself
	for _, lp := range r.Info.Depends {
		if lp.List == "" && lp.Package == "" {
			resultErr = multierror.Append(resultErr, fmt.Errorf("A list must depend on either a list or a package"))
			continue
		}
		if lp.List != "" && lp.Package != "" {
			resultErr = multierror.Append(resultErr, fmt.Errorf("On dependency of a list cannot be both a package and a list"))
			continue
		}
		if lp.List != "" {
			repos, resultErr = pm.addRepositoryWithExtends(index, lp.List, repos, resultErr)
		}
		if lp.Package != "" {
			// TODO: Implement package depends
			resultErr = multierror.Append(resultErr, fmt.Errorf("Package dependencies on lists are not implemented yet"))
			continue
		}
	}

	return repos, resultErr
}

func (pm *PackageManager) AddRepository(index, url string) error {

	repos := []*Repository{}
	resultErr := &multierror.Error{}
	repos, resultErr = pm.addRepositoryWithExtends(index, url, repos, resultErr)

	// Reverse the list of repos
	// See: https://stackoverflow.com/a/19239850
	for i, j := 0, len(repos)-1; i < j; i, j = i+1, j-1 {
		repos[i], repos[j] = repos[j], repos[i]
	}

	// Now extend PM's list of Repos with them
	pm.repos = append(pm.repos, repos...)

	return resultErr.ErrorOrNil()
}

func (rh *PackageManager) GetRepositories() []*Repository {
	return rh.repos
}

func NewPackageManager() (*PackageManager, error) {

	dpm := &PackageManager{
		repos:     []*Repository{},
		packages:  make(map[string]*RPackage),
		providers: make(map[string]map[string]*RPackage),
	}

	return dpm, nil
}

func (pm *PackageManager) Validate() error {
	rErr := &multierror.Error{}

	// Create a list of packages and providers
	for _, r := range pm.repos {
		for i := range r.Packages {
			p := r.Packages[i]
			// Packages are allowed to overwrite previous versions
			// if _, ok := pm.packages[p.Name]; ok {
			// 	rErr = multierror.Append(rErr, fmt.Errorf("%v: Package %v exists more than once", r.url, p.Name))
			// }
			pm.packages[p.Name] = &p
			if len(p.Provides) > 0 {
				for _, pn := range p.Provides {
					if _, ok := pm.providers[pn]; !ok {
						pm.providers[pn] = make(map[string]*RPackage)
					}
					pm.providers[pn][p.Name] = &p
				}
			}
		}
	}

	// Check Dependencies
	for _, p := range pm.packages {
		for _, d := range p.Dependencies {
			if prov, ok := pm.providers[d.Package]; ok {
				if d.Default != "" {
					if _, ok := prov[d.Default]; !ok {
						rErr = multierror.Append(rErr, fmt.Errorf("%v: Unknown default dependency package \"%s\" for package \"%s\"", p.Repository.GetList(), d.Default, p.Name))
					}
				}
				continue
			}

			if _, ok := pm.packages[d.Package]; ok {
				continue
			}

			rErr = multierror.Append(rErr, fmt.Errorf("%v: Unknown dependency \"%s\" for package \"%s\"", p.Repository.GetList(), d.Package, p.Name))
		}
	}

	// Check depends on self
	for _, p := range pm.packages {
		for _, d := range p.Dependencies {
			if d.Package == p.Name || d.Default == p.Name {
				rErr = multierror.Append(rErr, fmt.Errorf("%v: Package \"%s\" depends on itself", p.Repository.GetList(), p.Name))
			}
			for _, pn := range p.Provides {
				if pn == p.Name {
					rErr = multierror.Append(rErr, fmt.Errorf("%v: Package \"%s\" depends on/provides itself", p.Repository.GetList(), p.Name))
				}
			}
		}
	}

	return rErr.ErrorOrNil()
}

func (pm *PackageManager) getDependencies(inPackages []*RPackage, kpn map[string]int, rErr *multierror.Error) ([]*RPackage, map[string]int, *multierror.Error) {
	myPkgs := []*RPackage{}

	for _, p := range inPackages {
		for _, d := range p.Dependencies {
			// Check if already known
			if _, ok := kpn[d.Package]; ok {
				// Package or Provider already known
				continue
			}

			// Check if its a provider package
			if _, ok := pm.providers[d.Package]; ok {
				// Check if has default
				if d.Default != "" {
					// Default already known
					if _, ok := kpn[d.Default]; ok {
						continue
					}

					// Check if default is a known package
					dp, ok := pm.packages[d.Default]
					if !ok {
						rErr = multierror.Append(rErr, fmt.Errorf("Default package %v is unknown to me", d.Default))
						continue
					}

					// We know the default, check if it is known
					if _, ok := kpn[dp.Name]; !ok {
						// And its not already known
						myPkgs = append(myPkgs, dp)
					}
					kpn[dp.Name] = 0
					for _, prov := range dp.Provides {
						kpn[prov] = 0
					}

					continue
				}
			}

			// Check if a known package
			dp, ok := pm.packages[d.Package]
			if !ok {
				rErr = multierror.Append(rErr, fmt.Errorf("Package %v is unknown to me", d.Package))
				continue
			}
			// We know the package add it
			if _, ok := kpn[dp.Name]; !ok {
				myPkgs = append(myPkgs, dp)
			}
			kpn[dp.Name] = 0
			for _, prov := range dp.Provides {
				kpn[prov] = 0
			}
		}
	}

	if len(myPkgs) > 0 {
		myPkgs, kpn, rErr = pm.getDependencies(myPkgs, kpn, rErr)
	}

	inPackages = append(inPackages, myPkgs...)

	return inPackages, kpn, rErr
}

func (pm *PackageManager) GetDependencies(from []string) ([]*RPackage, *multierror.Error) {
	resultPackages := []*RPackage{}
	names := make(map[string]int)
	resultErr := &multierror.Error{}

	for _, myDep := range from {
		p, ok := pm.packages[myDep]
		if !ok {
			resultErr = multierror.Append(resultErr, fmt.Errorf("Package %v is unknown to me", myDep))
			continue
		}

		if _, ok := names[p.Name]; ok {
			resultErr = multierror.Append(resultErr, warning.Wrap(fmt.Errorf("Package %v is already known", p.Name)))
			continue
		}
		names[p.Name] = 0

		known := []string{}
		for _, prov := range p.Provides {
			if _, ok := names[prov]; ok {
				known = append(known, prov)
				continue
			}
			names[prov] = 0
		}
		if len(known) > 0 {
			resultErr = multierror.Append(resultErr, warning.Wrap(fmt.Errorf("Provider/s %v of package %v is/are already known", known, p.Name)))
			continue
		}

		resultPackages = append(resultPackages, p)
	}

	resultPackages, names, resultErr = pm.getDependencies(resultPackages, names, resultErr)

	return resultPackages, resultErr
}
