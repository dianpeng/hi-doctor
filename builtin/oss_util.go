package builtin

import (
	"github.com/dianpeng/hi-doctor/dvar"

	"fmt"
)

// helper functions for working with oss related functions

func ossGetPath(
	jobName string,
	localPath *dvar.DVar,
	env *dvar.EvalEnv,
) (string, error) {
	var out string

	if v, err := localPath.Value(env); err != nil {
		return "", fmt.Errorf("%s path execution failed: %s", jobName, err)
	} else {
		maybePath := v.String()
		if maybePath == "" {
			if ossPath, ok := env.Get("target", "oss_path"); !ok {
				return "", fmt.Errorf("%s path is empty, not specified in task", jobName)
			} else {
				out = ossPath.String()
			}
		} else {
			out = maybePath
		}
		return out, nil
	}
}

func ossGetObject(
	jobName string,
	localObject *dvar.DVar,
	env *dvar.EvalEnv,
) (string, error) {
	getObjectValue := func(v dvar.Val) string {
		return v.String()
	}

	// 2) value
	if v, err := localObject.Value(env); err != nil {
		return "", fmt.Errorf("%s object execution failed: %s", jobName, err)
	} else {
		maybeBody := getObjectValue(v)
		if maybeBody == "" {
			if ossBodyStr, ok := env.Get("target", "oss_body"); !ok {
				return "", fmt.Errorf(
					"%s body is empty, not specified in task and target",
					jobName,
				)
			} else {
				// the ossBody field is a string from the json, and it can include code
				if ossBodyExpr, err := dvar.NewDVarStringContext(
					ossBodyStr.String(),
				); err != nil {
					return "", fmt.Errorf("%s target.body compiles failed: %s", jobName, err)
				} else {
					if ossBody, err := ossBodyExpr.Value(env); err != nil {
						return "", fmt.Errorf("%s target.body execution failed: %s", jobName, err)
					} else {
						maybeBody = getObjectValue(ossBody)
					}
				}
			}
		}
		return maybeBody, nil
	}
}
