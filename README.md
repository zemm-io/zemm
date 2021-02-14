# zemm

Zemm - the dependency Management Tool for docker

## The Problem

### Unable to run additional plugins during runtime

While creating "the next big thing" a microcontainer application which should be as extensible as some PHP Frameworks in the wild,
the creator of zemm faced the problem that there are no tools which allow you to start microservice during runtime.

### No Tools for Lists of compatible Software

Another problem he faced was that there are no tools that create Lists of Software which is compatible with each other, like "generic" Pluginstores.

### Example compose File

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

## An example local override File

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

## Authors

René Jochum - rene@jochum.dev

## Maintainer

René Jochum - rene@jochum.dev

## License

Apache-2.0 License
