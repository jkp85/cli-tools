[![slack in](https://slackin-pypmyuhqds.now.sh/badge.svg)](https://slackin-pypmyuhqds.now.sh/)

# Command Line Interface (CLI) Tools

Command Line Interface (CLI) tools used to manage 3Blades resources. CLI tools connect to [3Blades backend API](https://github.com/3blades/app-backend).

## Local compilation

Install Go as instructed here [https://golang.org/doc/install](https://golang.org/doc/install).

Make sure you have go available in your shell.

	go version

To compile run:

	CGO_ENABLED=0 go build -o tbs .

Check if everything is ok:

	./tbs

If you have no error message then you are good to go :)
