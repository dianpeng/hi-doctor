package s14y

type Job struct {
	Name string
	Md5  string
}

type NewJob struct {
	Name   string
	Md5    string
	Data   string
	Origin string
	Delete bool // whether delete this file
}

type Config map[string]interface{}

type S14y interface {
	Refresh([]Job) []NewJob
}

func (c *Config) GetStringField(name string) string {
	v, ok := (*c)[name]
	if !ok {
		return ""
	}
	vv, ok := v.(string)
	if !ok {
		return ""
	}
	return vv
}

func (c *Config) GetName() string {
	return c.GetStringField("name")
}

func (c *Config) GetDir() string {
	return c.GetStringField("dir")
}

func (c *Config) GetCron() string {
	return c.GetStringField("cron")
}

func NewS14y(cfg Config) S14y {
	name := cfg.GetName()
	if name == "dir" {
		return newDir(cfg.GetDir())
	}
	if name == "oss" {
		return newOss(cfg)
	}
	return nil
}
