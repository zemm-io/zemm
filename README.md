# zemm

I'm the zemm commandline tool.

## The compose File

zemm-compose.yaml

```yaml
version: 1.0

# Both are the default no need to define them
indexes:
  zemm: https://hub.zemm.org
  docker: https://hub.docker.com

lists:
  # Import a list from a Team, this list also defines supported packages/lists
  main: minadmin/minadmin/1.0.0
  additional: []

# Packages that will be installed from the list(s) above
install:
  # Depends on lots of framework "tuatzemm" packages which are importet in lists/main
  - package: minadmin/minadmin_pgsql
    # This allows the package above to overlay config files of all packages
    overlay: true
```

## An example local File

zemm-compose.local.yaml

```yaml
version: 1.0

clear:
  lists_additional: False
  install: False

repositories:
  zemm: ../../repo/
  docker: https://registry.fk.jochum.dev

lists:
  # Overwrite Package abac_pgsql
  additional:
    - package: zemmaschaffa/abac_pgsql@6.6.6
```

## Commands

### zemm install

Will download all lists and theier dependencies, create a list of packages to install and download them.

### zemm docker up -d

Creates a docker-compose.yaml and runs "docker-compose up -d"

### zemm docker down

- Reverse, means App first then deps of the app then deps of the deps :)

### zemm docker ps

- Gives a list of running containers
