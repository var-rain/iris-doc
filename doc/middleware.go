package doc

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

/* 32 MB in memory max */
const MaxInMemoryMultipartSize = 32000000

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
	"Accept-Encoding":   false,
	"Accept-Language":   false,
	"Cache-Control":     false,
	"Connection":        false,
	"Origin":            false,
	"User-Agent":        false,
}

type Handler struct {
	nextHandler http.Handler
}

func Handle(nextHandler http.Handler) http.Handler {
	return &Handler{nextHandler: nextHandler}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !IsOn() {
		h.nextHandler.ServeHTTP(w, r)
		return
	}
	writer := NewResponseRecorder(w)
	call := Call{}
	Before(&call, r)
	h.nextHandler.ServeHTTP(writer, r)
	After(&call, writer, r)
}

func HandleFunc(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsOn() {
			next(w, r)
			return
		}
		call := Call{}
		writer := NewResponseRecorder(w)
		Before(&call, r)
		next(writer, r)
		After(&call, writer, r)
	})
}

func Before(call *Call, req *http.Request) {
	call.RequestHeader = ReadHeaders(req)
	call.RequestUrlParams = ReadQueryParams(req)
	val, ok := call.RequestHeader["Content-Type"]
	log.Println(val)
	if ok {
		ct := strings.TrimSpace(call.RequestHeader["Content-Type"])
		switch ct {
		case "application/x-www-form-urlencoded":
			fallthrough
		case "application/json, application/x-www-form-urlencoded":
			log.Println("Reading form")
			call.PostForm = ReadPostForm(req)
		case "application/json":
			log.Println("Reading body")
			call.RequestBody = *ReadBody(req)
		default:
			if strings.Contains(ct, "multipart/form-data") {
				handleMultipart(call, req)
			} else {
				log.Println("Reading body")
				call.RequestBody = *ReadBody(req)
			}
		}
	}
}

func ReadQueryParams(req *http.Request) map[string]string {
	params := map[string]string{}
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		return params
	}
	for k, v := range u.Query() {
		if len(v) < 1 {
			continue
		}
		params[k] = v[0]
	}
	return params
}

func handleMultipart(call *Call, req *http.Request) {
	call.RequestHeader["Content-Type"] = "multipart/form-data"
	req.ParseMultipartForm(MaxInMemoryMultipartSize)
	call.PostForm = ReadMultiPostForm(req.MultipartForm)
}

func ReadMultiPostForm(mpForm *multipart.Form) map[string]string {
	postForm := map[string]string{}
	for key, val := range mpForm.Value {
		postForm[key] = val[0]
	}
	return postForm
}

func ReadPostForm(req *http.Request) map[string]string {
	postForm := map[string]string{}
	for _, param := range strings.Split(*ReadBody(req), "&") {
		value := strings.Split(param, "=")
		postForm[value[0]] = value[1]
	}
	return postForm
}

func ReadHeaders(req *http.Request) map[string]string {
	b := bytes.NewBuffer([]byte(""))
	err := req.Header.WriteSubset(b, reqWriteExcludeHeaderDump)
	if err != nil {
		return map[string]string{}
	}
	headers := map[string]string{}
	for _, header := range strings.Split(b.String(), "\n") {
		values := strings.Split(header, ":")
		if strings.EqualFold(values[0], "") {
			continue
		}
		headers[values[0]] = values[1]
	}
	return headers
}

func ReadHeadersFromResponse(header http.Header) map[string]string {
	headers := map[string]string{}
	for k, v := range header {
		log.Println(k, v)
		headers[k] = strings.Join(v, " ")
	}
	return headers
}

func ReadBody(req *http.Request) *string {
	save := req.Body
	var err error
	if req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil
		}
	}
	b := bytes.NewBuffer([]byte(""))
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if req.Body == nil {
		return nil
	}
	var dest io.Writer = b
	if chunked {
		dest = httputil.NewChunkedWriter(dest)
	}
	_, err = io.Copy(dest, req.Body)
	if chunked {
		dest.(io.Closer).Close()
	}
	req.Body = save
	body := b.String()
	return &body
}

func After(call *Call, record *responseRecorder, r *http.Request) {
	if strings.Contains(r.RequestURI, ".ico") || !IsOn() {
		return
	}
	call.MethodType = r.Method
	call.CurrentPath = r.URL.Path
	call.ResponseBody = record.Body.String()
	call.ResponseCode = record.Status
	call.ResponseHeader = ReadHeadersFromResponse(record.Header())
	if IsStatusCodeValid(record.Status) {
		go GenerateHtml(call)
	}
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
