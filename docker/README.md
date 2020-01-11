# Build with docker

Use the `do.sh` script file in the top directory.

## Build the docker image

```shell script
./do.sh docker-image
```

## Build binary

```shell script
# installs dependencies, only once needed
./do.sh docker-dep 

# build the binary
./do.sh docker-build
```
