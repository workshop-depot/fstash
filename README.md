[![Go Report Card](https://goreportcard.com/badge/github.com/dc0d/fstash)](https://goreportcard.com/report/github.com/dc0d/fstash)

# fstash
stash a file or a tree of files for later reuse - a bit like `git stash`

# test

Use this command:

```
$ go test -v *.go
```

# usage story

Assume you write various applications and for each new project you add some initial files as the beginning skeleton. Also you replace some strings or expand some templates to add various information to the application, like author or other metadata.

This is one the of best fitting scenarios for using _fstash_. You create a stash from the skeleton files and then expand it whenever you are creating a new project. Also it is possible to provide the metadata to be injected into files.

As an example take a look at `sample-stash` directory. In this directory there is a Go file named `variables.go`. This file contains some variables that will be filled at compile time by Go compiler and come from `build.sh` file. Also there are two other variables `Author` and `License` which will be filled when we expand this stash for a new project.

First let’s create the stash:

```
$ cd sample-stash/
$ fstash create -n newproject
```

Now let’s create a new project which skeleton will be created from that stash:

```
$ cd ~/Documents/
$ mkdir newapp
$ cd newapp/
$ fstash expand -n newproject variables='{"Author":"Kaveh","License":"MIT"}'
```

Last command expands the stash we created in previous step, into current directory. The part `variables='{"Author":"Kaveh","License":"MIT"}'` indicates that `variables.go` is a (Go) text template file and the provided JSON should be passed to it as model data.

Now the content of variables.go is:

```golang
package main

// build flags
var (
        BuildTime  string
        CommitHash string
        GoVersion  string
        GitTag     string
)

// package info
const (
        Author  = "Kaveh"
        License = "MIT"
)
```

Tada! :)

I hope you find this tool useful.

