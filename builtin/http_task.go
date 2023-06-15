package builtin

import (
	"github.com/dianpeng/hi-doctor/check"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
	"github.com/dianpeng/hi-doctor/util"

	"github.com/mitchellh/mapstructure"

	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// http task definition, ie result of compilation from task.http definition
type httpTaskTemplate struct {
	name   string
	scheme dvar.DVar
	method dvar.DVar
	path   dvar.DVar
	header map[string]dvar.DVar
	body   dvar.DVar
	host   dvar.DVar
	close  bool

	timeout int64

	// --------------------------------------------------------------------------
	// TODO(dpeng): TLS related stuff, if needed
	// --------------------------------------------------------------------------

	check check.Check
}

func (h *httpTaskTemplate) Description() string {
	return fmt.Sprintf("http_task(%s)", h.name)
}

func (h *httpTaskTemplate) GenTask(env *dvar.EvalEnv) (task.TaskList, error) {
	return task.TaskList{
		newHttpTask(h),
	}, nil
}

type httpTaskDefine struct {
	Timeout int64             `mapstructure:"timeout"`
	Name    string            `mapstructure:"name"`
	Scheme  string            `mapstructure:"scheme"`
	Method  string            `mapstructure:"method"`
	Path    string            `mapstructure:"path"`
	Header  map[string]string `mapstructure:"header"`
	Body    string            `mapstructure:"body"`
	Host    string            `mapstructure:"host"`
	Close   bool              `mapstructure:"close"`
}

type httpTaskResultTLS struct {
	Version            string `json:"tls_version"`
	CipherSuite        string `json:"tls_cipher"`
	NegotiatedProtocol string `json:"tls_negotiated_proto"`
}

type httpTaskResult struct {
	ReqMethod string      `json:"req_method"`
	ReqUrl    string      `json:"req_url"`
	ReqBody   string      `json:"req_body"`
	ReqScheme string      `json:"req_scheme"`
	ReqIp     string      `json:"req_ip"`
	ReqPort   uint16      `json:"req_port"`
	ReqPath   string      `json:"req_path"`
	ReqHeader http.Header `json:"req_header"`

	RespOK     bool        `json:"resp_ok"`
	RespError  string      `json:"resp_error"`
	RespStatus int         `json:"resp_status"`
	RespHeader http.Header `json:"resp_header"`
	RespBody   string      `json:"resp_body"`
	RespProto  string      `json:"resp_proto"`

	// TLS related information
	RespIsTLS bool              `json:"resp_is_tls"`
	RespTLS   httpTaskResultTLS `json:"resp_tls"`

	// time statistics, need more ??
	Timestamp int64 `json:"timestamp"`
	RespTTFB  int64 `json:"resp_ttfb"`
	RespRT    int64 `json:"resp_rt"`
}

func populateHttpTaskDefine(opt spec.TaskOption,
) (*httpTaskDefine, error) {
	o := httpTaskDefine{
		Header: make(map[string]string),
	}

	err := mapstructure.Decode(opt, &o)
	if err != nil {
		return nil, fmt.Errorf("http_task, invalid Option input: %s", err)
	}
	return &o, nil
}

// compile http task template
func compileHttpTaskTemplate(m *httpTaskDefine,
	checkModel *spec.Check,
) (*httpTaskTemplate, error) {
	o := &httpTaskTemplate{
		header: make(map[string]dvar.DVar),
	}

	o.timeout = m.Timeout
	o.close = m.Close
	o.name = m.Name

	// http.Scheme
	if dv, err := dvar.NewDVarStringContext(m.Scheme); err != nil {
		return nil, fmt.Errorf("http_task.Scheme compile failed: %s", err)
	} else {
		o.scheme = dv
	}

	// http.Method
	if dv, err := dvar.NewDVarStringContext(m.Method); err != nil {
		return nil, fmt.Errorf("http_task.Method compile failed: %s", err)
	} else {
		o.method = dv
	}

	// http.Path
	if dv, err := dvar.NewDVarStringContext(m.Path); err != nil {
		return nil, fmt.Errorf("http_task.Path compile failed: %s", err)
	} else {
		o.path = dv
	}

	// http.Host
	if dv, err := dvar.NewDVarStringContext(m.Host); err != nil {
		return nil, fmt.Errorf("http_task.Host compile failed: %s", err)
	} else {
		o.host = dv
	}

	// http.Header
	for k, v := range m.Header {
		if dv, err := dvar.NewDVarStringContext(v); err != nil {
			return nil, fmt.Errorf("http_task.Header[%s] compile failed: %s", k, err)
		} else {
			o.header[k] = dv
		}
	}

	// http.Body
	if dv, err := dvar.NewDVarStringContext(m.Body); err != nil {
		return nil, fmt.Errorf("http_task.Body compile failed: %s", err)
	} else {
		o.body = dv
	}

	if ck, err := check.CompileCheck(checkModel); err != nil {
		return nil, fmt.Errorf("http_task.Check compile failed: %s", err)
	} else {
		o.check = ck
	}

	return o, nil
}

type httpTaskFactory struct{}

func (f *httpTaskFactory) SanityCheck(spec.TaskOption) error {
	return nil
}

func (f *httpTaskFactory) Compile(x spec.TaskOption, c *spec.Check) (task.TaskPlanner, error) {
	define, err := populateHttpTaskDefine(x)
	if err != nil {
		return nil, err
	}
	return compileHttpTaskTemplate(define, c)
}

func init() {
	task.RegisterTaskFactory("http", &httpTaskFactory{})
}

type httpTask struct {
	t *httpTaskTemplate

	// populate by Prepare function
	ip      string
	port    uint16
	method  string
	path    string
	header  http.Header
	body    string
	host    string
	isHttps bool
}

func newHttpTask(t *httpTaskTemplate) *httpTask {
	return &httpTask{
		t:      t,
		header: make(http.Header),
	}
}

func (h *httpTask) Description() string {
	return fmt.Sprintf("http_task[%s:%d]", h.ip, h.port)
}

// Sanity check for task operation. Used during parsing
func checkHttpMethod(method string) bool {
	if method == "GET" || method == "HEAD" || method == "POST" ||
		method == "PUT" || method == "PATCH" || method == "DELETE" ||
		method == "CONNECT" || method == "OPTIONS" || method == "TRACE" {
		return true
	}
	return false
}

func checkHttpPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] != '/' {
		return false
	}
	return true
}

// implementation of exec interface
func (h *httpTask) Prepare(env *dvar.EvalEnv) error {
	{
		var ip string
		if ipVal, hasIpVal := env.Get("target", "ip"); !hasIpVal {
			return fmt.Errorf("http_task.IP is not available")
		} else {
			ip = ipVal.String()
		}
		h.ip = ip
	}

	hostV := env.GetDef("target", "hostname", dvar.NewStringVal(""))
	hostHint := hostV.String()
	schemeHint := env.GetDef("target", "scheme", dvar.NewStringVal(""))

	if vv, err := h.t.method.Value(env); err != nil {
		return fmt.Errorf("http_task.Method execution failed: %s", err)
	} else {
		h.method = vv.String()
		if !checkHttpMethod(h.method) {
			return fmt.Errorf("http_task.Method %s invalid", h.method)
		}
	}

	if vv, err := h.t.path.Value(env); err != nil {
		return fmt.Errorf("http_task.Path execution failed: %s", err)
	} else {
		h.path = vv.String()
		if !checkHttpPath(h.path) {
			return fmt.Errorf("http_task.Path %s invalid path", h.path)
		}
	}

	if vv, err := h.t.body.Value(env); err != nil {
		return fmt.Errorf("http_task.Body execution failed: %s", err)
	} else {
		h.body = vv.String()
	}

	if vv, err := h.t.scheme.Value(env); err != nil {
		return fmt.Errorf("http_task.Scheme execution failed: %s", err)
	} else {
		isHttps := func(scheme string) bool {
			switch scheme {
			case "https":
				return true
			default:
				return false
			}
		}
		if scheme := vv.String(); scheme == "" {
			h.isHttps = isHttps(schemeHint.String())
		} else {
			h.isHttps = isHttps(scheme)
		}
	}

	// Port # detection.
	// The logic is as following, if the task defines a port #, then it will take
	// priority, otherwise check the target.port is defined or not, if both are
	// not defined, then based on http/https scheme to choose which port will be
	// used
	{
		var port uint16
		xx := env.GetOrNull("target", "port")
		if pv, ok := xx.Port(); ok {
			port = pv
		} else {
			// if we do not have port defined inside of target.port, then based on
			// scheme to choose a default port
			if h.isHttps {
				port = 443
			} else {
				port = 80
			}
		}
		h.port = port
	}

	// foreach header
	for k, v := range h.t.header {
		if vv, err := v.Value(env); err != nil {
			return fmt.Errorf("http_task.Header[%s] execution failed: %s", k, err)
		} else {
			h.header.Add(k, vv.String())
		}
	}

	// set host if needed

	// 1) If the task has host setup, ie does host overwrite, then it takes effect
	// 2) Otherwise, check whether there's a hostname hint during the target
	//    materialization, if so, use host hint, otherwise just leave as it is
	var hostVar string
	if vv, err := h.t.host.Value(env); err != nil {
		return fmt.Errorf("http_task.Host execution failed: %s", err)
	} else {
		hostVar = vv.String()
	}

	if hostVar != "" {
		h.header.Set("host", hostVar)
		h.host = hostVar
	} else if hostHint != "" {
		if h.header.Get("host") == "" {
			h.host = hostHint
		} else {
			h.host = h.header.Get("host")
		}
	} else {
		// check whether user have already setup host or not, if not uses ip:port
		// as hostname
		if h.header.Get("host") == "" {
			h.host = fmt.Sprintf("%s:%d", h.ip, h.port)
		} else {
			h.host = h.header.Get("host")
		}
	}

	return nil
}

func (h *httpTask) taskName() string {
	return h.t.name
}

func (h *httpTask) bodyReader() io.Reader {
	return strings.NewReader(h.body)
}

// run the task
func (h *httpTask) doRunHttp(env *dvar.EvalEnv) (*httpTaskResult, error) {
	out := &httpTaskResult{}

	// perform the http task requests and return everything into the global table
	client := &http.Client{
		Timeout: time.Duration(h.t.timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	var scheme string
	if h.isHttps {
		scheme = "https"
	} else {
		scheme = "http"
	}

	url := fmt.Sprintf("%s://%s:%d%s",
		scheme, // scheme of request
		h.ip,   // ip address
		h.port, // port number
		h.path, // path
	)

	req, err := http.NewRequest(h.method, url, h.bodyReader())
	if err != nil {
		return nil, fmt.Errorf("http_task, cannot create request: %s", url, err)
	}

	req.Close = h.t.close

	req.Header = h.header
	req.Host = h.host

	respStatusCode := 0
	respHeader := make(http.Header)
	respBody := ""
	respError := ""
	respHasError := false
	respProto := ""
	respIsTls := false

	respTlsVer := ""
	respTlsCipher := ""
	respTlsNProto := ""

	httpReqTs := time.Now().UnixMilli()
	resp, err := client.Do(req)
	httpRespTs := time.Now().UnixMilli()

	var httpBodyTs int64

	if err != nil {
		respStatusCode = 0
		respError = fmt.Sprintf("%s", err)
		respHasError = true
	} else {
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("http_task, client response body failed to read: %s", err)
		}
		httpBodyTs = time.Now().UnixMilli()
		respBody = string(data)
		respHeader = resp.Header
		respStatusCode = resp.StatusCode

		respIsTls = resp.TLS != nil
		if resp.TLS != nil {
			respTlsVer = util.GetTLSVersionName(resp.TLS.Version)
			respTlsCipher = util.GetTLSCipherSuitesName(resp.TLS.CipherSuite)
			respTlsNProto = resp.TLS.NegotiatedProtocol
		}
	}

	out.ReqMethod = h.method
	out.ReqUrl = url
	out.ReqBody = h.body
	out.ReqScheme = scheme
	out.ReqIp = h.ip
	out.ReqPort = h.port
	out.ReqPath = h.path
	out.ReqHeader = h.header

	out.RespOK = !respHasError
	out.RespError = respError
	out.RespStatus = respStatusCode
	out.RespHeader = respHeader
	out.RespBody = respBody
	out.RespProto = respProto

	out.RespTTFB = (httpRespTs - httpReqTs)
	out.RespRT = (httpBodyTs - httpReqTs)
	out.Timestamp = httpReqTs

	// TLS related stuff
	out.RespIsTLS = respIsTls
	out.RespTLS.Version = respTlsVer
	out.RespTLS.CipherSuite = respTlsCipher
	out.RespTLS.NegotiatedProtocol = respTlsNProto

	return out, nil
}

// FIXME(dpeng): this is slow, but it works and also it does not need any
//
//	thirdparty dependency, we can optionally bailout to a specific library
//	to do so either, though not now
func (h *httpTask) populateResult(env *dvar.EvalEnv, r *httpTaskResult) error {
	stat := util.ToMapInterface(r)
	env.RecordHistoricalResult("http", stat)
	return nil
}

func (h *httpTask) runHttp(env *dvar.EvalEnv) error {
	stat, err := h.doRunHttp(env)
	if err != nil {
		return err
	}
	return h.populateResult(env, stat)
}

func (h *httpTask) runCheck(env *dvar.EvalEnv) error {
	return h.t.check.Run(env)
}

func (h *httpTask) Run(env *dvar.EvalEnv) error {
	if err := h.runHttp(env); err != nil {
		return err
	}
	if err := h.runCheck(env); err != nil {
		return err
	}
	return nil
}
