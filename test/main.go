package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/exec"
	"github.com/dianpeng/hi-doctor/run"
	"github.com/dianpeng/hi-doctor/trigger"

	// for side-effect
	_ "github.com/dianpeng/hi-doctor/assert"
)

// test extension, help script to notify us that a test case is done
type testfactory struct{}
type testInfo map[string]bool

var (
	testinfo  = make(testInfo)
	assetsMap = make(dvar.ValMap)
)

func (t *testfactory) Create(e *exec.Executor) exec.Extension {
	b := e.Blackboard
	b["test"] = testinfo

	lib := make(map[string]interface{})
	lib["Done"] = func(name string, result bool) bool {
		testinfo[name] = result
		return result
	}
	return exec.Extension{
		Name:    "test",
		Inline:  true,
		Library: lib,
	}
}

func (t *testfactory) Description() string {
	return "test"
}

// test driver main. All it does is just to iterate through all the *.yaml from
// test folder and run them one by one
func runFile(f string) error {
	_, err := run.RunInspectionFile(assetsMap, f)
	return err
}

func runTestFolder(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return dir, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(dir, f.Name())
		testinfo["file://"+path] = false
		fmt.Printf("start to run test cases: %s\n", path)
		if err := runFile(path); err != nil {
			return path, err
		}
	}

	return "", nil
}

func printTestResult() {
	idx := 0
	for x, y := range testinfo {
		d := ""
		if y {
			d = "passed"
		} else {
			d = "not passed"
		}
		fmt.Printf("%d(%s)] %s\n", idx, x, d)
		idx++
	}
}

func init() {
	exec.AddExtension("test", &testfactory{})
}

var singleTest = flag.String("test", "", "specify single test file path")

func main() {
	assetsMap["asset1"] = dvar.NewStringVal("value")
	assetsMap["asset2"] = dvar.NewIntVal(1)
	assetsMap["asset3"] = dvar.NewBooleanVal(true)

	flag.Parse()

	if *singleTest != "" {
		path := *singleTest
		testinfo["file://"+path] = false
		if err := runFile(path); err != nil {
			fmt.Printf("execution error(%s): %s", path, err)
		}
	} else {
		x, err := runTestFolder("./test/cases")
		if err != nil {
			fmt.Printf("execution error(%s): %s", x, err)
		}
	}
	trigger.StopSafely()
	printTestResult()
}
