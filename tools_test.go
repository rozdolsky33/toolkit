package toolkit

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestNew(t *testing.T) {
	tools := New()
	if tools.MaxXMLSize != defaultMaxUpload {
		t.Errorf("Wrong MaxSize")
	}
}

func TestTools_PushJSONToRemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test Request Parameters
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}
	})
	var testTools Tools
	var foo struct {
		Bar string `json:"bar"`
	}
	foo.Bar = "bar"

	_, _, err := testTools.PushJSONToRemote("http://example.com/some/path", foo, client)
	if err != nil {
		t.Errorf("failed to call remote url: %s", err)
	}
}

// TestTools_RandomString verifies that the RandomString method returns a string of the correct length.
func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

// uploadTests is a slice of test cases for upload functionality including test name, allowed file types, renaming flag, and error expectation.
var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, errorExpected: true},
}

// TestTools_UploadFiles tests the file upload functionality via multipart form-data with various scenarios and configurations.
func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		//set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()
			// crete the form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")

			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()
		// read from the pipe which receives data
		request := httptest.NewRequest(http.MethodPost, "/upload", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}
		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("file %s not uploaded %s", e.name, err.Error())

			}
			// Clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}
		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}
		wg.Wait()
	}
}

// TestTools_UploadOneFile tests the UploadOneFile method to ensure a file can be uploaded, stored, and verified correctly.
func TestTools_UploadOneFile(t *testing.T) {
	//set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		defer writer.Close()

		// crete the form data field 'file'
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Error(err)
		}

		f, err := os.Open("./testdata/img.png")

		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("error decoding image", err)
		}
		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}
	}()
	// read from the pipe which receives data
	request := httptest.NewRequest(http.MethodPost, "/upload", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)

	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("file not uploaded %s", err.Error())
	}
	// Clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExist("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExist("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/myDir")

}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "Hello World", expected: "hello-world", errorExpected: false},
	{name: "empty string", s: "", expected: "", errorExpected: true},
	{name: "complex string", s: "Now is the time for all GOOD men! + fish & such &^123", expected: "now-is-the-time-for-all-good-men-fish-such-123", errorExpected: false},
	{name: " japanese string", s: "こんにちは世界", expected: "", errorExpected: true},
	{name: " japanese string and roman characters", s: "hello world こんにちは世界", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools

	for _, test := range slugTests {
		slug, err := testTools.Slugify(test.s)
		if err != nil && !test.errorExpected {
			t.Errorf("%s: error received but none expected: %s", test.name, err.Error())
		}

		if !test.errorExpected && slug != test.expected {
			t.Errorf("%s: wrong slug retrned; expected %s but got %s", test.name, test.expected, slug)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	var testTools Tools
	testTools.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "puppy.jpg")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "98827" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}
	if res.Header["Content-Disposition"][0] != "attachment; filename=\"puppy.jpg\"" {
		t.Error("wrong content disposition of", res.Header["Content-Disposition"][0])
	}
	_, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Error(err)
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{name: "valid json", json: `{"foo": "bar"}`, errorExpected: false, maxSize: 1024, allowUnknown: false},
	{name: "badly formatted json", json: `{"foo": }`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type", json: `{"foo": 1}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "two json files", json: `{"foo": "1""}{"alpha" : "beta"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "empty json", json: ``, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "syntax error in json", json: `{"foo": 1""`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown field in json", json: `{"fod": "1"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "allow unknown field in json", json: `{"fooo": "1"}`, errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "missing field name in json", json: `{jack: "1"}`, errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: " file too large", json: `{"foo"": "bar"}`, errorExpected: true, maxSize: 4, allowUnknown: true},
	{name: "not jason ", json: `Hello World!`, errorExpected: true, maxSize: 1024, allowUnknown: true},
}

func TestTools_ReadJSON(t *testing.T) {
	var testTools Tools
	for _, test := range jsonTests {
		// set the max file size
		testTools.MaxJSONSize = test.maxSize

		// allow/disallow unknown fields
		testTools.AllowUnknownFields = test.allowUnknown

		var decodedJSON struct {
			Foo string `json:"foo"`
		}

		// create a request with the body
		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(test.json)))
		if err != nil {
			t.Log("Error:", err)
		}

		// create a recorder
		rr := httptest.NewRecorder()

		err = testTools.ReadJSON(rr, req, &decodedJSON)

		if test.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none reived: %s", test.name, err.Error())
		}

		if !test.errorExpected && err != nil {
			t.Errorf("%s: error not expected but one received", test.name)
		}

		req.Body.Close()

	}
}

func TestTools_WriteJSON(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	payload := JSONResponse{
		Error:   false,
		Message: "foo",
	}
	headers := make(http.Header)
	headers.Add("FOO", "BAR")

	err := testTools.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write JSON: %v", err)
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	err := testTools.ErrorJSON(rr, errors.New("some error"), http.StatusServiceUnavailable)
	if err != nil {
		t.Errorf("failed to write JSON: %v", err)
	}
	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		t.Errorf("failed to read JSON: %v", err)
	}
	if !payload.Error {
		t.Errorf("error set to false in JSON, and it should be true")
	}
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("error code set to false in JSON, and it should be http.StatusServiceUnavailable")
	}
}

var writeXMLTests = []struct {
	name          string
	payload       any
	errorExpected bool
}{
	{
		name: "valid",
		payload: XMLResponse{
			Error:   false,
			Message: "foo",
		},
		errorExpected: false,
	},
	{
		name:          "invalid",
		payload:       make(chan int),
		errorExpected: true,
	},
}

func TestTools_WriteXML(t *testing.T) {

	for _, e := range writeXMLTests {
		// create a variable of type toolkit.Tools, and just use the defaults.
		var testTools Tools

		rr := httptest.NewRecorder()

		header := make(http.Header)
		header.Add("FOO", "BAR")
		err := testTools.WriteXML(rr, http.StatusOK, e.payload, header)

		if err != nil && !e.errorExpected {
			t.Errorf("%s, failed to write XML: %v", e.name, err)
		}

		if err == nil && e.errorExpected {
			t.Errorf("%s: error expected, but none reived", e.name)
		}

	}
}

var xmlTests = []struct {
	name          string
	xml           string
	maxBytes      int
	errorExpected bool
}{
	{
		name:          "Good XML",
		xml:           `<?xml version="1.0" encoding="UTF-8"?><note><to>John Smith</to><from>Jane Jones</from></note>`,
		errorExpected: false,
	},
	{
		name:          "Badly formated XML",
		xml:           `<?xml version="1.0" encoding="UTF-8"?><note><xx>John Smith</to><from>Jane Jones</from></note>`,
		errorExpected: true,
	},
	{
		name:          "Too Big Size",
		xml:           `<?xml version="1.0" encoding="UTF-8"?><note><to>John Smith</to><from>Jane Jones</from></note>`,
		maxBytes:      10,
		errorExpected: true,
	},
	{
		name: "Double XML",
		xml: `<?xml version="1.0" encoding="UTF-8"?><note><to>John Smith</to><from>Jane Jones</from></note>
						<?xml version="1.0" encoding="UTF-8"?><note><to>Luke Skywalker</to><from>R2D2</from></note>`,
		errorExpected: true,
	},
}

func TestTools_ReadXML(t *testing.T) {
	for _, e := range xmlTests {

		var tools Tools

		if e.maxBytes != 0 {
			tools.MaxXMLSize = e.maxBytes
		}

		// create a request with the body.
		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(e.xml)))
		if err != nil {
			t.Log("Error:", err)
		}

		// create a test response recorder, which satisfies the requirements for a ResponseWriter.
		rr := httptest.NewRecorder()

		// call ReadXML and check for an error
		var note struct {
			To   string `xml:"to"`
			From string `xml:"from"`
		}

		err = tools.ReadXML(rr, req, &note)
		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none reived", e.name)
		} else if !e.errorExpected && err != nil {
			t.Errorf("%s: error not expected but one received", e.name)
		}
	}
}

func TestTools_ErrorXML(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	err := testTools.ErrorXML(rr, errors.New("some error"), http.StatusServiceUnavailable)
	if err != nil {
		t.Errorf("failed to write XML: %v", err)
	}

	var requestPayload XMLResponse
	decoder := xml.NewDecoder(rr.Body)
	err = decoder.Decode(&requestPayload)
	if err != nil {
		t.Errorf("failed to read XML: %v", err)
	}

	if !requestPayload.Error {
		t.Errorf("error set to false in XML, and it should be true")
	}

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("wrong status code returned; expected 503, but got %d", rr.Code)
	}
}
