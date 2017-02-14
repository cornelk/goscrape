goscrape is a web scraper built with Golang.

To install:

```
go install github.com/cornelk/goscrape
```

Usage:
```
goscrape http://website
```

Advantages over existing tools like wget, httrack, Teleport Pro:
* Free and open source
* Available for all platforms that Golang supports
* Excluding URLS will not fetch them (unlike [wget](http://savannah.gnu.org/bugs/?20808))

Planned features:

* Downsizing images
* Excluding of files
* Excluding of URLS
