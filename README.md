# Toolkit Package

[![CI Build](https://github.com/rozdolsky33/toolkit/actions/workflows/build.yml/badge.svg)](https://github.com/rozdolsky33/toolkit/actions/workflows/build.yml)



This toolkit package provides a set of utilities designed for the Go programming language to handle common tasks like generating random strings and handling file uploads in web applications.

## Features

The included tools are:

- [ ] Read JSON
- [ ] Write JSON
- [ ] Produce a JSON encoded error response
- [X] Upload files via HTTP requests with optional renaming and file type validation.
- [X] Download a static file
- [X] Get a random string of length n
- [ ] Post JSON to a remote service
- [X] Create a directory, including all parent directories, if it does not already exist
- [X] Create a URL safe slug from a string

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
t := toolkit.Tools{
    MaxFileSize:      1024 * 1024 * 1024, // 1 GB
    AllowedFileTypes: []string{"image/jpeg", "image/png"},
}
```

### Generating Random Strings

Use the `RandomString` method to generate a random string of specified length:

```go
randomStr := t.RandomString(16)
fmt.Println("Random String:", randomStr)
```

### Uploading Files

Use the `UploadFiles` method to handle file uploads from HTTP requests. You can also specify whether to rename the uploaded files or not:

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    uploadDir := "./uploads"

    uploadedFiles, err := t.UploadFiles(r, uploadDir, true)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    for _, file := range uploadedFiles {
        fmt.Fprintf(w, "Uploaded File: %s (original: %s), Size: %d bytes\n",
            file.NameFileName, file.OriginalFileName, file.FileSize)
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

The `Tools` struct is used to instantiate the toolkit. This struct holds configuration for file uploads.

- `MaxFileSize int`: Maximum allowed file size for uploads (in bytes).
- `AllowedFileTypes []string`: List of allowed file MIME types for validation.

### UploadedFile

The `UploadedFile` struct holds information about the uploaded files.

- `NameFileName string`: The name of the file saved on the server.
- `OriginalFileName string`: The original name of the uploaded file.
- `FileSize int64`: The size of the uploaded file in bytes.

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