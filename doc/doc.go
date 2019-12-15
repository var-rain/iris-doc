package doc

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
)

var count int
var config *Config

var spec = &Spec{}

func IsOn() bool {
	return config.On
}

func Init(conf *Config) {
	config = conf
	if conf.DocPath == "" {
		conf.DocPath = "apidoc.html"
	}

	filePath, err := filepath.Abs(conf.DocPath + ".json")
	dataFile, err := os.Open(filePath)
	defer dataFile.Close()
	if err == nil {
		_ = json.NewDecoder(io.Reader(dataFile)).Decode(spec)
		generateHtml()
	}
}

func add(x, y int) int {
	return x + y
}

func mult(x, y int) int {
	return (x + 1) * y
}

func GenerateHtml(call *Call) {
	shouldAddPathSpec := true
	for k, apiSpec := range spec.ApiSpecs {
		if apiSpec.Path == call.CurrentPath && apiSpec.HttpVerb == call.MethodType {
			shouldAddPathSpec = false
			call.Id = count
			count += 1
			deleteCommonHeaders(call)
			avoid := false
			for _, currentApiCall := range spec.ApiSpecs[k].Calls {
				if call.RequestBody == currentApiCall.RequestBody &&
					call.ResponseCode == currentApiCall.ResponseCode &&
					call.ResponseBody == currentApiCall.ResponseBody {
					avoid = true
				}
			}
			if !avoid {
				spec.ApiSpecs[k].Calls = append(apiSpec.Calls, *call)
			}
		}
	}

	if shouldAddPathSpec {
		apiSpec := Api{
			HttpVerb: call.MethodType,
			Path:     call.CurrentPath,
		}
		call.Id = count
		count += 1
		deleteCommonHeaders(call)
		apiSpec.Calls = append(apiSpec.Calls, *call)
		spec.ApiSpecs = append(spec.ApiSpecs, apiSpec)
	}
	filePath, err := filepath.Abs(config.DocPath)
	dataFile, err := os.Create(filePath + ".json")
	if err != nil {
		log.Println(err)
		return
	}
	defer dataFile.Close()
	data, err := json.Marshal(spec)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = dataFile.Write(data)
	if err != nil {
		log.Println(err)
		return
	}
	generateHtml()
}

func generateHtml() {
	funcs := template.FuncMap{"add": add, "mult": mult}
	t := template.New("API Documentation").Funcs(funcs)
	htmlString := Template
	t, err := t.Parse(htmlString)
	if err != nil {
		log.Println(err)
		return
	}
	filePath, err := filepath.Abs(config.DocPath)
	if err != nil {
		panic("Error while creating file path : " + err.Error())
	}
	homeHtmlFile, err := os.Create(filePath)
	defer homeHtmlFile.Close()
	if err != nil {
		panic("Error while creating documentation file : " + err.Error())
	}
	homeWriter := io.Writer(homeHtmlFile)
	_ = t.Execute(homeWriter, map[string]interface{}{"array": spec.ApiSpecs,
		"baseUrls": config.BaseUrls, "Title": config.DocTitle})
}

func deleteCommonHeaders(call *Call) {
	delete(call.RequestHeader, "Accept")
	delete(call.RequestHeader, "Accept-Encoding")
	delete(call.RequestHeader, "Accept-Language")
	delete(call.RequestHeader, "Cache-Control")
	delete(call.RequestHeader, "Connection")
	delete(call.RequestHeader, "Cookie")
	delete(call.RequestHeader, "Origin")
	delete(call.RequestHeader, "User-Agent")
}

func IsStatusCodeValid(code int) bool {
	if code >= 200 && code < 300 {
		return true
	} else {
		return false
	}
}
