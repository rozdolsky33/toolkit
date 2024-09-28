# Toolkit Package

[![CI Build](https://github.com/rozdolsky33/toolkit/actions/workflows/build.yml/badge.svg)](https://github.com/rozdolsky33/toolkit/actions/workflows/build.yml)
[![Coverage Status](https://coveralls.io/repos/github/rozdolsky33/toolkit/badge.svg?branch=main)](https://coveralls.io/github/rozdolsky33/toolkit?branch=main)

This toolkit package provides a set of utilities designed for the Go programming language to handle common tasks like generating random strings and handling file uploads in web applications.

## Features

The included tools are:

- [X] Read JSON
- [X] Write JSON
- [X] Produce a JSON encoded error response
- [X] Upload files via HTTP requests with optional renaming and file type validation.
- [X] Download a static file
- [X] Get a random string of length n
- [X] Post JSON to a remote service
- [X] Create a directory, including all parent directories, if it does not already exist
- [X] Create a URL-safe slug from a string

## Installation

To install the package, use the following command:

```sh
go get github.com/rozdolsky33/toolkit
```

## Usage

### Importing the Package

Before using the package, import it in your Go project:

```go
import "github.com/rozdolsky33/toolkit"
```

### Initializing the Tools

Create an instance of the `Tools` type to access its methods:

```go
tools := toolkit.Tools{
    MaxFileSize:      1024 * 1024 * 1024, // 1 GB
    AllowedFileTypes: []string{"image/jpeg", "image/png"},
}
```

### Generating Random Strings

Use the `RandomString` method to generate a random string of specified length:

```go
randomStr := tools.RandomString(16)
fmt.Println("Random String:", randomStr)
```

### Uploading Files

Use the `UploadFiles` method to handle file uploads from HTTP requests. You can also specify whether to rename the uploaded files or not:

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    uploadDir := "./uploads"

    uploadedFiles, err := tools.UploadFiles(r, uploadDir, true)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    for _, file := range uploadedFiles {
        fmt.Fprintf(w, "Uploaded File: %s (original: %s), Size: %d bytes\n",
            file.NewFileName, file.OriginalFileName, file.FileSize)
    }
}
```

In your main function, set up the HTTP server to use this handler:

```go
http.HandleFunc("/upload", uploadHandler)
http.ListenAndServe(":8080, nil)
```

## Structs

### Tools

The `Tools` struct is used to instantiate the toolkit. This struct holds configuration for file uploads and JSON operations.

- `MaxFileSize int`: Maximum allowed file size for uploads (in bytes).
- `AllowedFileTypes []string`: List of allowed file MIME types for validation.
- `MaxJSONSize int`: Maximum allowed JSON size in bytes.
- `AllowUnknownFields bool`: Flag to allow unknown JSON fields.

### UploadedFile

The `UploadedFile` struct holds information about the uploaded files.

- `NewFileName string`: The name of the file saved on the server.
- `OriginalFileName string`: The original name of the uploaded file.
- `FileSize int64`: The size of the uploaded file in bytes.

### JSONResponse

The `JSONResponse` struct is used to format JSON responses.

- `Error bool`: Indicates if the response is an error.
- `Message string`: The message to be included in the response.
- `Data interface{}`: Optional data payload.

## Methods

### `RandomString`

Generates a random string of specified length `n`.

```go
func (t *Tools) RandomString(n int) string
```

### `UploadFiles`

Handles file uploads from HTTP requests, validates file type and optionally renames files.

```go
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error)
```

- `r *http.Request`: The HTTP request object.
- `uploadDir string`: Directory path where files will be uploaded.
- `rename ...bool`: Optional boolean to specify whether to rename uploaded files.

### `CreateDirIfNotExist`

Creates a directory if it does not exist.

```go
func (t *Tools) CreateDirIfNotExist(dir string) error
```

- `dir string`: The directory path.

### `Slugify`

Transforms an input string into a URL-friendly slug.

```go
func (t *Tools) Slugify(s string) (string, error)
```

- `s string`: The input string to be slugified.

### `DownloadStaticFile`

Downloads a file and tries to force the browser to avoid displaying it in the browser window by setting content disposition.

```go
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string)
```

- `w http.ResponseWriter`: The HTTP response writer.
- `r *http.Request`: The HTTP request.
- `p string`: The file path.
- `file string`: The filename.
- `displayName string`: The display name.

### `ReadJSON`

Reads and decodes JSON from a request body.

```go
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error
```

- `w http.ResponseWriter`: The HTTP response writer.
- `r *http.Request`: The HTTP request.
- `data interface{}`: The target data structure.

### `WriteJSON`

Encodes data as JSON and writes it to the response.

```go
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error
```

- `w http.ResponseWriter`: The HTTP response writer.
- `status int`: The HTTP status code.
- `data interface{}`: The payload to be encoded as JSON.
- `headers ...http.Header`: Optional headers.

### `ErrorJSON`

Generates and sends a JSON error response.

```go
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error
```

- `w http.ResponseWriter`: The HTTP response writer.
- `err error`: The error to be included in the response.
- `status ...int`: Optional HTTP status code.

### `PushJSONToRemote`

Sends the given data as a JSON payload to a specified URI via HTTP POST using an optional custom HTTP client.

```go
func (t *Tools) PushJSONToRemote(uri string, data interface{}, client ...*http.Client) (*http.Response, int, error)
```

- `uri string`: The target URI.
- `data interface{}`: The data to be sent as JSON.
- `client ...*http.Client`: Optional custom HTTP client.

# MIT License

### Copyright (c) 2024 Volodymyr Rozdolsky

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.