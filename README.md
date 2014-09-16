# git-trunk #

Your ultimate Trunk Based Development (TBD) CLI utility.

Actually, I don't know about you, but we use it here at [Salsita](https://www.salsitasoft.com/).

## Instalation (script) ##
Run
```
wget --no-check-certificate -q -O - https://github.com/salsita/gitflow2/raw/develop/bash/install.sh | sudo bash
```
This will create a few files and a directory in your `cwd` so we recommend
doing it in `/tmp` or similar.

## Installation (manual) ##

1. Install [Go](https://golang.org/dl/) (used Go 1.3.1, but any Go 1.x should do the trick).
2. Set up a Go [workspace](https://golang.org/doc/code.html#Workspaces).
3. Run `go get github.com/tchap/git-trunk`. This will get the sources and install the executable into the workspace.
   Add `$GOPATH/bin` into `PATH` to be able to run the executable directly from the command line.
4. Run `go build` in `bin/hooks/commit-msg`, then use it as a Git [hook](http://git-scm.com/book/en/Customizing-Git-Git-Hooks) in your repo.
5. Copy files from `bash/` directory into your `${PATH}`.

## System Requirements ##

* `git>=2.0.0` in your `PATH`

## License ##

`MIT`, see the `LICENSE` file.
