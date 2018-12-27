package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/alecthomas/kingpin"
)

func main() {
	switch kingpin.Parse() {
	case "create":
		if *createStashContent == "." {
			*createStashContent = _wd
		}
		if err := createStash(*createStashName, *createStashContent, _appHome); err != nil {
			fmt.Println(err)
			return
		}
	case "pop":
		if *popDstDir == "." {
			*popDstDir = _wd
		}
		if err := popStash(*popStashName, _appHome, *popDstDir); err != nil {
			fmt.Println(err)
			return
		}
	case "list":
		l, err := listDepth(_appHome, 5)
		if err != nil {
			fmt.Println(err)
			return
		}
		var items []interface{}
		for _, v := range l {
			items = append(items, v)
		}
		fmt.Println(items...)
	}
}

var (
	createCommand      = kingpin.Command("create", "creating stash based on the content of a directory")
	createStashName    = createCommand.Flag("stash-name", "name of this stash, lower case, only numbers, alphabet and - and _").Short('n').Required().String()
	createStashContent = createCommand.Flag("stash-content", "the directory that its content will be used to create the stash").Short('c').Default(".").String()

	popCommand   = kingpin.Command("pop", "pop stash and expand it into a directory")
	popStashName = popCommand.Flag("stash-name", "name of this stash, lower case, only numbers, alphabet and - and _").Short('n').Required().String()
	popDstDir    = popCommand.Flag("destination", "the directory that its content will be expanded to").Short('d').Default(".").String()

	listCommand = kingpin.Command("list", "lists existing file stashes")
)

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	_appHome = filepath.Join(usr.HomeDir, ".fstash")
	if err := os.MkdirAll(_appHome, 0777); err != nil {
		panic(err)
	}

	_wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	kingpin.CommandLine.HelpFlag.Short('h')
}

var (
	_appHome string
	_wd      string
)
