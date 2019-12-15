package doc

type Call struct {
	Id                   int
	CurrentPath          string
	MethodType           string
	PostForm             map[string]string
	RequestHeader        map[string]string
	CommonRequestHeaders map[string]string
	ResponseHeader       map[string]string
	RequestUrlParams     map[string]string
	RequestBody          string
	ResponseBody         string
	ResponseCode         int
}

type Api struct {
	HttpVerb string
	Path     string
	Calls    []Call
}

type Config struct {
	On       bool
	BaseUrls map[string]string
	DocTitle string
	DocPath  string
}

type Spec struct {
	ApiSpecs []Api
}
