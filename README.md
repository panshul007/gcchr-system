# GCCHR System

This is a monorepo for all GCCHR System code.  Every project in the GCCHR landscape is a sub directory.

## Setup

### Install Docker Engine
Download and install native docker engine for your OS from the official [docker website](https://store.docker.com/editions/community/docker-ce-desktop-mac)

### Setup and run mongo db
Use the following command to run mongodb as a docker container. This will workon MacOSX environment:
```bash
mkdir -p ~/Development/data/mongo

sudo docker run -d -p 27017:27017 -v ~/Development/data/mongo:/data/db --name mongodb mongo

```
The first command will make the directory for storing the database, so that even if the mongodb container is restarted,
the data will not be lost.

The second command downloads the latest mongodb container image and runs it with the name `mongodb`, exposing the default port for mongodb clients
to connect to and mounts the created data directory to the container for storage.

To check if the container is running:
```bash
docker ps
```
This should list all the runnning containers on the system. You should see a container with name `mongodb` in the list.

#### Restart mongo db
It for some reason, the mongo db container exits or needs to be restarted (like after a system reboot), use the following:
```bash
docker container start mongodb
``` 

If you get an exception with the above command about `container exists`, this is bug in the docker enginer, which will be
fixed in the next docker update.
Work around for the bug: remove the existing container (data will not be harmed), then recreate the container.

Remove a container:
```bash
docker rm mongodb
```

Recreate it with the above command: `sudo docker run...`

## Running the Gcchr-System Core

The system core can be run using a binary built for the platform or from the code. While the system is in development, 
we will only run using the code.

### Running from the code

For running the core from code , we need to have the following installed: 
* Golang - use the official Go website to install Go on ur system, for mac os, simply use `homebrew`
   * `brew install go`
   * configure the `$GOPATH` environment variable to a directory where you wish to store all the source code. eg: `~/Development/gospace`
   * create sub folder structure in ur go path:
      * `mkdir -p ~/Development/gospace/src`
      * `mkdir -p ~/Development/gospace/bin`
* dep - dependency manager for Go. 
   * refer to the [official page](https://golang.github.io/dep/docs/installation.html) for your platform.
   * for mac use `brew install -u dep`  

Next is getting the source code for the core. Clone the repo into the `$GOPATH/src/` folder.

After you clone the `gcchr-system` repository into the `$GOPATH/src/` folder, do the following to configure:

```bash
cd $GOPATH/src/gcchr-system
dep ensure
```
This will download all the dependencies required for running the system and put them in the `vendor` folder of the project.

Now run the core (from inside the $GOPATH/src/gcchr-system folder):
```bash
go run core/core.go
``` 

This will run the core server and create an Admin account. Please contact the author to get the default admin account credentials.

The server can be accessed at: `http://localhost:1986`

Please use the issues page on the repository to send feedback, issues or suggestions.