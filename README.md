# goscrape

A web scraper built with Golang.

## Install:

```
go install github.com/cornelk/goscrape
```

## Usage:
```
goscrape http://website
```

Advantages over existing tools like wget, httrack, Teleport Pro:
* Free and open source
* Available for all platforms that Golang supports
* Excluded URLS will not be fetched (unlike [wget](https://savannah.gnu.org/bugs/?20808))
* No incomplete temp files are left on disk
* Downloaded files are skipped in a new scraper run
* Sane default values

## Planned features:

* Downsizing images
* Excluding of files
* Excluding of URLS
* Concurrent downloads
