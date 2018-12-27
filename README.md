# fstash
stash a file or a tree of files for later reuse - a bit like `git stash`

# test

Use this command:

```
$ go test -v *.go
```

# usage

Currently only creating file stashes and expanding them using pop sub-command is implemented.

A directory tree can be used as a skeleton for a named stash. Also it can be expanded using the subcommand `pop` by providing the name of the stash.

The help for the app:

```
usage: fstash [<flags>] <command> [<args> ...]

Flags:
  -h, --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  create --stash-name=STASH-NAME [<flags>]
    creating stash based on the content of a directory

  pop --stash-name=STASH-NAME [<flags>]
    pop stash and expand it into a directory

```

Also the help for subcommands are provided too. For subcommand `create`:

```
usage: fstash create --stash-name=STASH-NAME [<flags>]

creating stash based on the content of a directory

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
  -n, --stash-name=STASH-NAME  name of this stash, lower case, only numbers, alphabet and - and _
  -c, --stash-content="."      the directory that its content will be used to create the stash

```

And for sumcommand `pop`:

```
usage: fstash pop --stash-name=STASH-NAME [<flags>]

pop stash and expand it into a directory

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
  -n, --stash-name=STASH-NAME  name of this stash, lower case, only numbers, alphabet and - and _
  -d, --destination="."        the directory that its content will be expanded to
```
