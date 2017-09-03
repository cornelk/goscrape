# goscrape [![Build Status](https://travis-ci.org/cornelk/goscrape.svg?branch=master)](https://travis-ci.org/cornelk/goscrape) [![GoDoc](https://godoc.org/github.com/cornelk/goscrape?status.svg)](https://godoc.org/github.com/cornelk/goscrape) [![Go Report Card](https://goreportcard.com/badge/cornelk/goscrape)](https://goreportcard.com/report/github.com/cornelk/goscrape) [![codecov](https://codecov.io/gh/cornelk/goscrape/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/goscrape)

A web scraper built with Golang. It downloads the content of a website or blog and allows you to read it offline.

## Install:

You need to have Golang installed, otherwise follow the guide at [https://golang.org/doc/install](https://golang.org/doc/install).

```
go install github.com/cornelk/goscrape
```

## Usage:
```
goscrape http://website.com
```

Features and advantages over existing tools like wget, httrack, Teleport Pro:
* Free and open source
* Available for all platforms that Golang supports
* JPEG and PNG images can be converted down in quality to save disk space
* Excluded URLS will not be fetched (unlike [wget](https://savannah.gnu.org/bugs/?20808))
* No incomplete temp files are left on disk
* Downloaded asset files are skipped in a new scraper run
* Assets from external domains are downloaded automatically
* Sane default values

## Options:

```
Scrape a website and create an offline browseable version on the disk

Usage:
  goscrape http://website.com [flags]

Flags:
      --config string         config file (default is $HOME/.goscrape.yaml)
  -d, --depth uint            download depth, 0 for unlimited (default 10)
  -x, --exclude stringArray   exclude URLs with PERL Regular Expressions support
  -i, --imagequality uint     image quality, 0 to disable reencoding
  -o, --output string         output directory to write files to
  -v, --verbose               verbose output
```
