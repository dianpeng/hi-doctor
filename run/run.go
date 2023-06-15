package run

// Run a single inspection cases, ie register it into the background trigger
// and waiting it to start to execute

import (
	"github.com/dianpeng/hi-doctor/dvar"   // global assets
	"github.com/dianpeng/hi-doctor/exec"   // plan execution
	"github.com/dianpeng/hi-doctor/loader" // mode loading
	"github.com/dianpeng/hi-doctor/plan"   // plan formation

	_ "github.com/dianpeng/hi-doctor/assert"
	_ "github.com/dianpeng/hi-doctor/builtin"
	_ "github.com/dianpeng/hi-doctor/metrics"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

func RunInspection(
	assets dvar.ValMap,
	yamlData string,
	context string,
) (*plan.Plan, error) {
	hash := md5.Sum([]byte(yamlData))
	md5 := hex.EncodeToString(hash[:])

	// 1) Start to loading YAML model
	model, err := loader.ParseData(yamlData)
	if err != nil {
		return nil, err
	}
	model.Info.Origin = context
	model.Info.Md5 = md5
	model.Info.Source = yamlData
	model.Info.Timestamp = time.Now()

	// 2) Plan compilation
	plan, err := plan.Compile(model)
	if err != nil {
		return nil, err
	}

	// 3) Plan execution
	executor := exec.NewExecutor(assets, plan)
	if err := executor.Start(); err != nil {
		return nil, err
	} else {
		return plan, nil
	}
}

func RunInspectionFile(assets dvar.ValMap, path string) (*plan.Plan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return RunInspection(assets, string(data), fmt.Sprintf("file://%s", path))
}

func GetInspectionName(yaml string) (string, error) {
	model, err := loader.ParseData(yaml)
	if err != nil {
		return "", err
	}
	return model.Name, nil
}

func RunInspectionDebug(
	assets dvar.ValMap,
	yamlData string,
	context string,
) (*plan.Plan, error) {
	hash := md5.Sum([]byte(yamlData))
	md5 := hex.EncodeToString(hash[:])

	// 1) Start to loading YAML model
	model, err := loader.ParseData(yamlData)
	if err != nil {
		return nil, err
	}
	model.Info.Origin = context
	model.Info.Md5 = md5
	model.Info.Source = yamlData
	model.Info.Timestamp = time.Now()
	model.Trigger = "trigger.Now()"

	// 2) Plan compilation
	plan, err := plan.Compile(model)
	if err != nil {
		return nil, err
	}

	// 3) Plan execution
	executor := exec.NewExecutor(assets, plan)
	if err := executor.Start(); err != nil {
		return nil, err
	} else {
		return plan, nil
	}
}
