[![slack in](https://slackin-pypmyuhqds.now.sh/badge.svg)](https://slackin-pypmyuhqds.now.sh/)

# Command Line Interface (CLI) Tools

Command Line Interface (CLI) tools used to manage 3Blades resources. CLI tools connect to [3Blades backend API](https://github.com/3blades/app-backend).

Review our [online documentation](https://docs.3blades.io) for a full list of available command options.

## Local compilation

Install Go as instructed here [https://golang.org/doc/install](https://golang.org/doc/install). Although described in the installation instructions, the basic steps to set up Go are described below.

Although not obligatory, its a good idea to set your GOPATH working directory:

    mkdir $HOME/go
    GOPATH=$HOME/go

Make sure you have go available in your shell.

    go version

If you don't see any output, make sure your path reflects:

    PATH=$PATH:/usr/local/go/bin

Get all you need and compile:

    go get github.com/3Blades/cli-tools/tbs

If you added your bin folder from GOPATH to your PATH then you can just run:

    tbs

If don't then:

    cd $GOPATH/bin
    ./tbs

If you have no error message then you are good to go :)

If you want to recompile from local source then:

    cd $GOPATH/src/github.com/3Blades/cli-tools/tbs
    go install

To update cli-tools you can do:

    go get -u github.com/3Blades/cli-tools/tbs

## Config

In order for cli-tools to work with [3Blades API server](https://github.com/3blades/app-backend) you need to put your api endpoint to config file.
CLI are looking for config file in your home directory. Default config file can be json, yaml or toml for example

	.threeblades.yaml

Currently supported options are:

	root: localhost:5000 // api root
	namespace: [your_username]
