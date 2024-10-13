package toolkit

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// randomStringSource defines the character set used for generating random strings.
const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVXWYZ0123456789_+"

// defaultMaxUpload the default max upload size (10 mb)
const defaultMaxUpload = 10485760

// Tools is the type used to instantiate this module. Any variable of this type will have access
// to all the methods with the receiver *Tools.
type Tools struct {
	MaxJSONSize        int         // maximum size of JSON file we'll process
	MaxXMLSize         int         // maximum size of XML file we'll process
	MaxFileSize        int         // maximum size of uploaded files in bytes
	AllowedFileTypes   []string    // allowed file types for upload (e.g. image/jpeg)
	AllowUnknownFields bool        // if set to true, allow unknown fields in JSON
	ErrorLog           *log.Logger // the info log.
	InfoLog            *log.Logger // the error log.
}

// JSONResponse is the type used for sending JSON around.
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// XMLResponse is the type used for sending XML around.
type XMLResponse struct {
	Error   bool        `xml:"error"`
	Message string      `xml:"message"`
	Data    interface{} `xml:"data,omitempty"`
}

// UploadedFile is a struct used to save information about an uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// New returns a new toolbox with sensible defaults.
func New() Tools {
	return Tools{
		MaxJSONSize: defaultMaxUpload,
		MaxXMLSize:  defaultMaxUpload,
		MaxFileSize: defaultMaxUpload,
		InfoLog:     log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:    log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// RandomString returns a string of random characters of length n, using randomStringSource as the source for the string
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// UploadOneFile uploads a single file from the provided HTTP request, storing it in the specified directory.
// If the optional rename argument is true or not provided, the file will be renamed.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)

	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles uploads one or more file to a specified directory, and gives the files a random name.
// It returns a slice containing the newly named files, the original file names, the size of the files,
// and potentially an error. If the optional last parameter is set to true, then we will not rename
// the files, but will use the original file names.
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	var uploadedFiles []*UploadedFile
	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024 // 1Gb
	}

	err := t.CreateDirIfNotExist(uploadDir)
	if err != nil {
		return nil, err
	}

	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("error parsing multipart form: " + err.Error())
	}
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buff := make([]byte, 512)
				_, err = infile.Read(buff)

				if err != nil {
					return nil, err
				}

				//TODO: Check to see if the file type is permitted
				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}
				if !allowed {
					return nil, errors.New("file type not allowed: " + fileType)
				}
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}
				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				uploadedFile.OriginalFileName = hdr.Filename

				var outFile *os.File
				defer outFile.Close()
				if outFile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outFile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}
				uploadedFiles = append(uploadedFiles, &uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}

// CreateDirIfNotExist creates a directory with the specified name if it does not already exist.
func (t *Tools) CreateDirIfNotExist(dir string) error {
	const mode = 0755
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// Slugify transforms an input string into a URL-friendly slug by replacing non-alphanumeric characters with hyphens.
func (t *Tools) Slugify(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string not permitted")
	}
	var regEx = regexp.MustCompile(`[^a-z\d]+`)

	slug := strings.Trim(regEx.ReplaceAllString(strings.ToLower(s), "-"), "-")

	if len(slug) == 0 {
		return "", errors.New("after removing characters, slug is zero length")
	}

	return slug, nil
}

// DownloadStaticFile downloads a file, and tries to force the browser to avoid displaying it
// in the browser window by setting content disposition. It also allows specification of the display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	fp := path.Join(p, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))
	http.ServeFile(w, r, fp)

}

// ReadJSON tries to read the body of a request and converts it from JSON to a variable. The third parameter, data,
// is expected to be a pointer, so that we can read data into it.
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {

	// Check content-type header; it should be application/json. If it's not specified,
	// try to decode the body anyway.
	if r.Header.Get("Content-Type") != "" {
		contentType := r.Header.Get("Content-Type")
		if strings.ToLower(contentType) != "application/json" {
			return errors.New("Content-Type must be application/json")
		}
	}

	// Set a sensible default for the maximum payload size.
	maxBytes := defaultMaxUpload

	// If MaxJSONSize is set, use the value instead of default.
	if t.MaxJSONSize != 0 {
		maxBytes = t.MaxJSONSize
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	if !t.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}

	// Attempt to decode the data, and figure out what the error is, if any, to send back a human-readable response
	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains icnorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains an invalid JSON (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("error unmarshalling JSON: %s", err.Error())

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only one JSON value")
	}

	return nil
}

// WriteJSON takes a response status code and arbitrary data and writes json to the client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// if we have a value as the last parameter in the function call, then we are setting a custom header.
	if len(headers) > 0 {
		for key, val := range headers[0] {
			w.Header()[key] = val
		}
	}

	// Set the content type and send response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil
}

// ErrorJSON takes an error, and optionally a status code, and generates and sends a JSON error message
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	// if a custom response code is specified, use that instead of bad request.
	if len(status) > 0 {
		statusCode = status[0]
	}

	// build JSON Payload.
	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return t.WriteJSON(w, statusCode, payload)
}

// PushJSONToRemote posts arbitrary json to some url, and returns the response, the response
// status code, and error, if any. The final parameter, client, is optional, and will default
// to the standard http.Client. It exists to make testing possible without an active remote
// url.
func (t *Tools) PushJSONToRemote(uri string, data interface{}, client ...*http.Client) (*http.Response, int, error) {
	// Create json
	jsonData, err := json.Marshal(data)

	if err != nil {
		return nil, 0, err
	}
	// Check for custom http client
	httpClient := &http.Client{}
	if len(client) > 0 {
		httpClient = client[0]
	}

	// Build the request and set the header
	request, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	// Call the remote uri
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	// Send response back
	return response, response.StatusCode, nil
}

// WriteXML takes a response status code and arbitrary data and writes an XML response to the client.
// The Content-Type header is set to application/xml
func (t *Tools) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.Marshal(data)
	if err != nil {
		return err
	}

	// If we have value as the last parameter in the function call, then we are setting a custom header.
	if len(headers) > 0 {
		for key, val := range headers[0] {
			w.Header()[key] = val
		}
	}
	// Set the content type and send response. According the RFC 7303, txt/xml and application/xml are to be treated as the same, so we'll just pick one.
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	// Add the XML header
	xmlOut := []byte(xml.Header + string(out))
	_, err = w.Write(xmlOut)
	if err != nil {
		return err
	}

	return nil
}

// ReadXML tries to read the body of an XML request into a variable. The third parameter, data, is expected be a pointer, so we can read data into it.
func (t *Tools) ReadXML(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := defaultMaxUpload

	// If MaxXMLSize is set, use that value instead of default
	if t.MaxXMLSize != 0 {
		maxBytes = t.MaxXMLSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := xml.NewDecoder(r.Body)

	// Attempt to decode the data.
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only one XML value")
	}
	return nil
}

// ErrorXML takes and error, and optionally a response status code, and generates adn sends an XML error response.

func (t *Tools) ErrorXML(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	// If a custom response code is specified, use that instead of bad request.
	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload XMLResponse
	payload.Error = true
	payload.Message = err.Error()

	return t.WriteXML(w, statusCode, payload)
}
