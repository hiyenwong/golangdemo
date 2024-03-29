Golint is a linter for Go source code.

[![Build Status](https://travis-ci.org/golang/lint.svg?branch=master)](https://travis-ci.org/golang/lint)

To install, run
  go get -u github.com/golang/lint/golint

Invoke golint with one or more filenames, a directory, or a package named
by its import path. Golint uses the same import path syntax as the go
command (https://golang.org/cmd/go/#hdr-Import_path_syntax) and therefore
also supports relative import paths like "./...". Additionally the "..."
wildcard can be used as suffix on relative and absolute file paths to recurse
into them.

The output of this tool is a list of suggestions in Vim quickfix format,
which is accepted by lots of different editors.

Golint differs from gofmt. Gofmt reformats Go source code, whereas
golint prints out style mistakes.

Golint differs from govet. Govet is concerned with correctness, whereas
golint is concerned with coding style. Golint is in use at Google, and it
seeks to match the accepted style of the open source Go project.

The suggestions made by golint are exactly that: suggestions.
Golint is not perfect, and has both false positives and false negatives.
Do not treat its output as a gold standard. We will not be adding pragmas
or other knobs to suppress specific warnings, so do not expect or require
code to be completely "lint-free".
In short, this tool is not, and will never be, trustworthy enough for its
suggestions to be enforced automatically, for example as part of a build process.

If you find an established style that is frequently violated, and which
you think golint could statically check, file an issue at
  https://github.com/golang/lint/issues


Contributions
-------------
Contributions to this project are welcome, though please send mail before
starting work on anything major. Contributors retain their copyright, so we
need you to fill out a short form before we can accept your contribution:
  https://developers.google.com/open-source/cla/individual


Vim
---
Add this to your ~/.vimrc:
  set rtp+=$GOPATH/src/github.com/golang/lint/misc/vim
If you have multiple entries in your GOPATH, replace $GOPATH with the right value.

Running :Lint will run golint on the current file and populate the quickfix list.

Optionally, add this to your ~/.vimrc to automatically run golint on :w
  autocmd BufWritePost,FileWritePost *.go execute 'Lint' | cwindow


Emacs
-----
Add this to your .emacs file:
  (add-to-list 'load-path (concat (getenv "GOPATH")  "/src/github.com/golang/lint/misc/emacs"))
  (require 'golint)
If you have multiple entries in your GOPATH, replace $GOPATH with the right value.

Running M-x golint will run golint on the current file.
For more usage, see Compilation-Mode:
  http://www.gnu.org/software/emacs/manual/html_node/emacs/Compilation-Mode.html
