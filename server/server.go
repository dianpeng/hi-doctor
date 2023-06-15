package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/metrics"
	"github.com/dianpeng/hi-doctor/plan"
	"github.com/dianpeng/hi-doctor/run"
	sd "github.com/dianpeng/hi-doctor/s14y"
	"github.com/dianpeng/hi-doctor/trigger"

	"github.com/julienschmidt/httprouter"
)

type server struct {
	assets dvar.ValMap
	jobs   map[string]*plan.Plan
	sync.Mutex
	sd sd.S14y
}

type jobVersion struct {
	Name        string `json:"name"`
	Md5Checksum string `json:"md5_checksum"`
	Origin      string `json:"origin"`
	Timestamp   string `json:"timestamp"`
}

var (
	theServer = newServer()
)

func newServer() *server {
	return &server{
		assets: make(dvar.ValMap),
		jobs:   make(map[string]*plan.Plan),
	}
}

func getReqContext(req *http.Request) string {
	reqId := req.Header.Get("x-request-id")
	addr := req.RemoteAddr
	xff := req.Header.Get("x-forwarded-for")
	ts := time.Now().UnixMilli()

	return fmt.Sprintf("http_post{req_id: %s, remote_addr: %s, xff: %s, ts: %d}",
		reqId,
		addr,
		xff,
		ts,
	)
}

func (s *server) onS14y() {
	s.Lock()
	defer s.Unlock()

	// 0) populate the current job list
	cur := []sd.Job{}
	for _, x := range s.jobs {
		cur = append(cur, sd.Job{
			Name: x.Name,
			Md5:  x.Info.Md5Checksum,
		})
	}

	// 1) run the refresh
	for _, entry := range s.sd.Refresh(cur) {
		if entry.Delete {
			if x, ok := s.jobs[entry.Name]; ok {
				x.Stop()
				delete(s.jobs, entry.Name)
			}
		} else {
			if err := s.addOrCreate(
				entry.Name,
				entry.Origin,
				entry.Data,
			); err != nil {
				log.Printf("service discovery refresh on %s, %s, %s failed: %s",
					entry.Name,
					entry.Origin,
					entry.Md5,
					err,
				)
			}
		}
	}
}

func (s *server) startS14y(cfg sd.Config) error {
	sd := sd.NewS14y(cfg)
	if sd == nil {
		return fmt.Errorf("invalid service discovery %s", cfg.GetName())
	}
	s.sd = sd
	s.onS14y()
	trigger.Cron(
		cfg.GetCron(),
		func() {
			s.onS14y()
		},
	)
	return nil
}

func (s *server) addOrCreate(
	name string,
	origin string,
	data string,
) error {
	if name == "" {
		n, err := run.GetInspectionName(data)
		if err != nil {
			return err
		}
		name = n
	}

	if oldPlan, ok := s.jobs[name]; ok {
		oldPlan.Stop()
		delete(s.jobs, name)
	}

	plan, err := run.RunInspection(
		s.assets,
		data,
		origin,
	)
	if err != nil {
		return err
	}
	if plan.Name != name {
		plan.Stop()
		return fmt.Errorf("job %s yaml definition name mismatch", name)
	}

	s.jobs[name] = plan
	return nil
}

func (s *server) add(
	name string,
	origin string,
	data string,
) error {
	if name == "" {
		n, err := run.GetInspectionName(data)
		if err != nil {
			return err
		}
		name = n
	}

	if _, ok := s.jobs[name]; ok {
		return fmt.Errorf("job %s already existed", name)
	}

	plan, err := run.RunInspection(
		s.assets,
		data,
		origin,
	)
	if err != nil {
		return err
	}
	if plan.Name != name {
		plan.Stop()
		return fmt.Errorf("job %s yaml definition name mismatch", name)
	}

	s.jobs[name] = plan
	return nil
}

// ----------------------------------------------------------------------------
// Basic CRUD interfaces

func onOverwrite(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("cannot read body %s", err)))
		return
	}

	if err := theServer.addOrCreate(
		"",
		getReqContext(req),
		string(data),
	); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("plan execution failed :%s", err)))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func onAdd(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("cannot read body %s", err)))
		return
	}

	if err := theServer.add(
		"",
		getReqContext(req),
		string(data),
	); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("plan execution failed :%s", err)))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func onRemove(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	if v, has := theServer.jobs[name]; has {
		v.Stop() // async stop
		delete(theServer.jobs, name)
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}
}

func onInfo(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	if v, has := theServer.jobs[name]; has {
		j, _ := json.MarshalIndent(v, "", "  ")
		w.WriteHeader(200)
		w.Write(j)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}
}

func onVersion(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	data := []jobVersion{}
	for _, v := range theServer.jobs {
		data = append(data, jobVersion{
			Name:        v.Name,
			Md5Checksum: v.Info.Md5Checksum,
			Origin:      v.Info.Origin,
			Timestamp:   v.Info.Timestamp.String(),
		})
	}

	j, _ := json.MarshalIndent(data, "", "  ")
	w.WriteHeader(200)
	w.Write(j)
}

func onList(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	data := []*plan.Plan{}
	for _, v := range theServer.jobs {
		data = append(data, v)
	}

	j, _ := json.MarshalIndent(data, "", "  ")
	w.WriteHeader(200)
	w.Write(j)
}

func StartServer(cfg sd.Config, assets dvar.ValMap, addr string) {
	theServer.assets = assets
	router := httprouter.New()
	if err := theServer.startS14y(cfg); err != nil {
		log.Printf("server service discovery %s\n", err)
		return
	}

	router.POST("/test/overwrite", onOverwrite)
	router.POST("/test/add", onAdd)
	router.GET("/test/remove/:name", onRemove)
	router.GET("/test/list", onList)
	router.GET("/test/info/:name", onInfo)
	router.GET("/test/version", onVersion)

	router.Handler(http.MethodGet, "/debug/pprof/:xxx", http.DefaultServeMux)
	router.Handler(http.MethodGet, "/prometheus", metrics.PrometheusHttpHandler())

	log.Printf("The *hi-doctor* server starts to listen on %s", addr)
	http.ListenAndServe(addr, router)
}
