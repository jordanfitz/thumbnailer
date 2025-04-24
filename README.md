[![Go Reference](https://pkg.go.dev/badge/github.com/jordanfitz/thumbnailer.svg)](https://pkg.go.dev/github.com/jordanfitz/thumbnailer)

## Thumbnailer

Thumbnailer is a simple utility library and CLI for generating image thumbnails, written in Golang.

It's heavily inspired by [prplecake/go-thumbnail](https://github.com/prplecake/go-thumbnail), with some additional functionality and adjusted ergonomics.

### Usage

#### As a library

To install Thumbnailer as a library, `go get` it:

```shell
go get github.com/jordanfitz/thumbnailer
```

Then, you might make use of the library like this:

```go
package main

import "github.com/jordanfitz/thumbnailer"

func main() {
    inputData, _ := os.ReadFile("input-image.jpg")

    t := thumbnailer.New().
        With(thumbnailer.OutFormat(thumbnailer.JPG)). // output a JPG
        With(thumbnailer.Quality(50)).                // encode the JPG with quality 50
        With(thumbnailer.MaxSize(200)).               // don't let the thumbnail exceed 200px, width or height
        With(thumbnailer.Image(inputData))            // use inputData as the source image

    outputData, _ := t.Create()

    // do something with outputData
    // ...
}
```

#### As a command-line utility

To install the CLI, run `go install`:

```shell
go install github.com/jordanfitz/thumbnailer/cmd/thumbnailer@latest
```

Then, use the utility like this:

```shell
# generate thumbnails for all PNG files in my-images/, outputting
# JPG files with a quality of 50 into the directory my-images/thumbs/
thumbnailer my-images/*.png -o jpg -q 50 -o my-images/thumbs
```

More information can be gleaned from `thumbnailer -h`.
