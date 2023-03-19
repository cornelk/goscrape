# goscrape - create offline browsable copies of websites

[![Build status](https://github.com/cornelk/goscrape/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/cornelk/goscrape/actions)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/cornelk/goscrape)
[![Go Report Card](https://goreportcard.com/badge/github.com/cornelk/goscrape)](https://goreportcard.com/report/github.com/cornelk/goscrape)
[![codecov](https://codecov.io/gh/cornelk/goscrape/branch/main/graph/badge.svg?token=NS5UY28V3A)](https://codecov.io/gh/cornelk/goscrape)

A web scraper built with Golang. It downloads the content of a website and allows it to be archived and read offline.

## Features

Features and advantages over existing tools like wget, httrack, Teleport Pro:

* Free and open source
* Available for all platforms that Golang supports
* JPEG and PNG images can be converted down in quality to save disk space
* Excluded URLS will not be fetched (unlike [wget](https://savannah.gnu.org/bugs/?20808))
* No incomplete temp files are left on disk
* Downloaded asset files are skipped in a new scraper run
* Assets from external domains are downloaded automatically
* Sane default values

## Limitations

* No GUI version, console only

## Installation

There are 2 options to install goscrape:

1. Download and unpack a binary release from [Releases](https://github.com/cornelk/goscrape/releases)

or

2. Compile the latest release from source:

```
go install github.com/cornelk/goscrape@latest
```

Compiling the tool from source code needs to have a recent version of [Golang](https://go.dev/) installed.

## Usage
```
goscrape http://website.com
```

## Options

```
Scrape a website and create an offline browsable version on the disk.

Usage: goscrape [--include INCLUDE] [--exclude EXCLUDE] [--output OUTPUT] [--depth DEPTH] [--imagequality IMAGEQUALITY] [--timeout TIMEOUT] [--proxy PROXY] [--user USER] [--useragent USERAGENT] [--verbose] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --include INCLUDE, -n INCLUDE
                         only include URLs with PERL Regular Expressions support
  --exclude EXCLUDE, -x EXCLUDE
                         exclude URLs with PERL Regular Expressions support
  --output OUTPUT, -o OUTPUT
                         output directory to write files to
  --depth DEPTH, -d DEPTH
                         download depth, 0 for unlimited [default: 10]
  --imagequality IMAGEQUALITY, -i IMAGEQUALITY
                         image quality, 0 to disable reencoding
  --timeout TIMEOUT, -t TIMEOUT
                         time limit in seconds for each HTTP request to connect and read the request body
  --proxy PROXY, -p PROXY
                         HTTP proxy to use for scraping
  --user USER, -u USER   user[:password] to use for authentication
  --useragent USERAGENT, -a USERAGENT 
                         user agent to use for scraping
  --verbose, -v          verbose output
  --help, -h             display this help and exit
```
