package cvvapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type httpResponseData struct {
	Status string
	StatusCode int
	Proto string
	ProtoMajor int
	ProtoMinor int
	Header http.Header
	ContentLength int64
	TransferEncoding []string
}
	
func newHttpResponseData(o *http.Response) httpResponseData {
	return httpResponseData{
		Status: o.Status,
		StatusCode: o.StatusCode,
		Proto: o.Proto,
		ProtoMajor: o.ProtoMajor,
		ProtoMinor: o.ProtoMinor,
		Header: o.Header,
		ContentLength: o.ContentLength,
		TransferEncoding: o.TransferEncoding,
	}
}

func (data *httpResponseData) toHttpResponse() *http.Response {

	return &http.Response{
		Status: data.Status,
		StatusCode: data.StatusCode,
		Proto: data.Proto,
		ProtoMajor: data.ProtoMajor,
		ProtoMinor: data.ProtoMinor,
		Header: data.Header,
		ContentLength: data.ContentLength,
		TransferEncoding: data.TransferEncoding,
	}
}

// Middleware to obtain a snapshot of a test run. It saves to disk the result of every http request, 
// hashing url and body, and reading the snapshotted response if a corresponding file is found 
func HttpSnapshoterBuilder(path string, printHash bool) HttpMiddleware {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return func(req *http.Request, next HttpHandler) (*http.Response, error) {
		hash := sha256.New()
		hash.Write([]byte(req.URL.String()))
		if req.Body != nil {
			body, err := req.GetBody()
			
			if err != nil {
				return nil, err
			}

			bodyBytes, err := io.ReadAll(body)
			
			if err != nil {
				return nil, err
			}
			
			hash.Write(bodyBytes)
			body.Close()
		}

		baseName := hex.EncodeToString(hash.Sum(nil))
		
		defer func () {
			if r := recover(); r != nil {
				fmt.Println(baseName)
				panic(r)
			}
		}()

		if _, err:= os.Stat(filepath.Join(path, baseName + ".header.txt")); os.IsNotExist(err) {
			resp, err := next(req)

			if err != nil {
				return nil, err
			}

			buffer := bytes.Buffer{}
			encoder := json.NewEncoder(&buffer)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(newHttpResponseData(resp))
			
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(filepath.Join(path, baseName + ".header.txt"), buffer.Bytes(), os.ModePerm)
			if err != nil {
				return nil, err
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			
			if err != nil {
				return nil, err
			}
			
			err = os.WriteFile(filepath.Join(path, baseName + ".body.txt"), []byte(body), os.ModePerm)
			if err != nil {
				return nil, err
			}

			resp.Body = io.NopCloser(bytes.NewReader(body))

			return resp, err
		}
		
		reqDataFile, err := os.Open(filepath.Join(path, baseName + ".header.txt"))
		if err != nil {
			return nil, err
		}
		defer reqDataFile.Close()

		var reqData = httpResponseData{}
		err = json.NewDecoder(reqDataFile).Decode(&reqData)
		if err != nil {
			return nil, err
		}

		bodyDataFile, err := os.Open(filepath.Join(path, baseName + ".body.txt"))
		if err != nil {
			return nil, err
		}

		cachedResp := reqData.toHttpResponse()
		cachedResp.Body = bodyDataFile

		if printHash {
			fmt.Println(baseName)
		}

		return cachedResp, nil
	}
}

// Creates a simple in memory cache for http requests hashing url and body. Usable for testing optimization purposes.
func MemoryCacheBuilder(expiration time.Duration) HttpMiddleware {
	type cacheEntry struct {
		resp httpResponseData
		body []byte
		age time.Time
	}
	var cache = make(map[string]cacheEntry)

	var cleanup func ()
	cleanup = func() {
		for key, val := range cache {
			if time.Since(val.age) > expiration {
				delete(cache, key)
			}
		}

		time.AfterFunc(expiration, func () {
			cleanup()
		})
	}

	cleanup()

	return func(req *http.Request, next HttpHandler) (*http.Response, error) {
		hash := sha256.New()
		hash.Write([]byte(req.URL.String()))
		if req.Body != nil {
			body, err := req.GetBody()
			
			if err != nil {
				return nil, err
			}

			bodyBytes, err := io.ReadAll(body)
			
			if err != nil {
				return nil, err
			}
			
			hash.Write(bodyBytes)
			body.Close()
		}

		reqHash := hex.EncodeToString(hash.Sum(nil))

		if entry, ok:= cache[reqHash]; !ok {
			resp, err := next(req)

			if err != nil {
				return nil, err
			}
			
			entry := cacheEntry{
				resp: newHttpResponseData(resp),
				age: time.Now(),
			}

			entry.body, err = io.ReadAll(resp.Body)
			resp.Body.Close()			
			
			if err != nil {
				return nil, err
			}

			resp.Body = io.NopCloser(bytes.NewReader(entry.body))

			cache[reqHash] = entry

			return resp, err
		} else {
	
			cachedResp := entry.resp.toHttpResponse()
			cachedResp.Body = io.NopCloser(bytes.NewReader(entry.body))
	
			return cachedResp, nil
		}
	}
}