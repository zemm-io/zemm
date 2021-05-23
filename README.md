# zemm

Zemm - the Dependency Management Tool for docker

## The problem

### Long list of services that need to work together

With microservices if you often have long lists of containers that should work together, the easiest way to distribute that is by using a docker-compose.yaml.
Now people gonna change your yaml for theier needs, how to upgrade that?

### Unable to run additional plugins during runtime

While creating "the next big thing" a microcontainer application which should be as extensible as some PHP Frameworks in the wild,
the creator of zemm faced the problem that there are no tools which allow you to start microservice during runtime.

### No Tools for Lists of compatible Software

Another problem he faced was that there are no tools that create Lists of Software which is compatible with each other, like "generic" Pluginstores.

## The solution

### Installable lists of compatible Software

Developers will have the ability to define lists of containers on our/own Servers with theier configuration in zemm and they also will have a solution to define "supported" plugins.

### Install additional plugins/microservices during runtime

There will be a container that runs zemm and has an api for other microservices to command it.

### Example zemm File

zemm.yaml

```yaml
version: 1.0

# Both are the default no need to define them
indexes:
  zemm: https://hub.zemm.io
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

local.zemm.yaml

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

### zemm registry list

List all available registries.

### zemm registry add

Add (another) registry.

### zemm registry remove

Remove a registry.

### zemm publish [registry]

Publish your package (the current directory containing a "zemmpkg.yaml") to the registry.

### zemm install

Will download all lists and theier dependencies, create a list of packages to install and download them.

### zemm compose up -d

Creates a docker-compose.yaml and runs "docker-compose up -d"

### zemm compose down

- Reverse, means App first then deps of the app then deps of the deps :)

### zemm compose ps

- Gives a list of running containers

### zemm update

Download the newest lists and packages of the same version from zemm.io

### zemm upgrade [-q] [-y]

Check if there are newer version's of the lists available and upgrade after confirmation

### zemm publish [-l hub.zemm.org]

Publish your package on a zemm server

## The name

The name Zemm comes from the Vorarlberger dialect and means "together".

## Authors

René Jochum - rene@jochum.dev

## Maintainer

René Jochum - rene@jochum.dev

## License

Apache-2.0 License
