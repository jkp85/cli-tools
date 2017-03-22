[![slack in](https://slackin-pypmyuhqds.now.sh/badge.svg)](https://slackin-pypmyuhqds.now.sh/)

# Command Line Interface (CLI) Tools

Command Line Interface (CLI) tools used to manage 3Blades resources. CLI tools connect to [3Blades backend API](https://github.com/3blades/app-backend).

## Local compilation

Install Go as instructed here [https://golang.org/doc/install](https://golang.org/doc/install).

Make sure you have go available in your shell.

	go version

Check if you have GOPATH in your env variables (default is $HOME/go)

It's good idea to add $GOPATH/bin folder to your PATH variable.

Get all you need:

	go get github.com/3Blades/cli-tools

If you added your bin folder from GOPATH to your PATH then you can just run:

	cli-tools

If don't then:

	cd $GOPATH/bin
	./cli-tools

If you have no error message then you are good to go :)

If you want to recompile it then:

	cd $GOPATH/src/github.com/3Blades/cli-tools
	go install


To update cli-tools you can do:

	go get -u github.com/3Blades/cli-tools

## Config

In order for cli-tools to work with your api server you need to put your api endpoint to config file.
CLI are looking for config file in your home directory. Default config file can be json, yaml or toml for example

	.threeblades.yaml

Currently supported options are:

	root: localhost:5000 // api root
	namespace: [your_username]
