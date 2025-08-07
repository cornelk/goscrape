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

Scrape a website by running
```
goscrape http://website.com
```

To serve the downloaded website directory in a local run webserver use
```
goscrape --serve website.com
```

## Options

```
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
  --serve SERVE, -s SERVE
                         serve the website using a webserver
  --serverport SERVERPORT, -r SERVERPORT
                         port to use for the webserver [default: 8080]
  --cookiefile COOKIEFILE, -c COOKIEFILE
                         file containing the cookie content
  --savecookiefile SAVECOOKIEFILE
                         file to save the cookie content
  --header HEADER, -h HEADER
                         HTTP header to use for scraping
  --proxy PROXY, -p PROXY
                         proxy to use in format scheme://[user:password@]host:port (supports HTTP, HTTPS, SOCKS5 protocols)
  --user USER, -u USER   user[:password] to use for HTTP authentication
  --useragent USERAGENT, -a USERAGENT
                         user agent to use for scraping
  --verbose, -v          verbose output
  --help, -h             display this help and exit
  --version              display version and exit
```

## Cookies

Cookies can be passed in a file using the `--cookiefile` parameter and a file containing
cookies in the following format:

```
[{"name":"user","value":"123"},{"name":"sessioe","value":"sid"}]
```

## Proxy Configuration

The `--proxy` flag supports multiple proxy protocols for scraping through different types of proxy servers:

### Supported Proxy Types

- **HTTP proxies**: `http://proxy.example.com:8080`
- **HTTPS proxies**: `https://proxy.example.com:8080`
- **SOCKS5 proxies**: `socks5://proxy.example.com:1080`

### Authentication

Proxies requiring authentication can include credentials in the URL:

```bash
# HTTP proxy with authentication
goscrape --proxy http://username:password@proxy.example.com:8080 http://website.com

# SOCKS5 proxy with authentication
goscrape --proxy socks5://user:pass@proxy.example.com:1080 http://website.com
```
