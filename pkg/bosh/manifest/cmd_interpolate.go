package manifest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SUSE/go-patch/patch"
	"github.com/pkg/errors"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

// InterpolateFromSecretMounts reads explicit secrets from a folder and writes an interpolated manifest to the output.json file in /mnt/quarks volume mount.
func InterpolateFromSecretMounts(boshManifestBytes []byte, variablesDir string, outputFilePath string) error {
	var vars []boshtpl.Variables

	variables, err := ioutil.ReadDir(variablesDir)
	if err != nil {
		return errors.Wrapf(err, "could not read variables directory")
	}

	for _, variable := range variables {
		// Each directory is a variable name
		if variable.IsDir() {
			staticVars := boshtpl.StaticVariables{}
			// Each filename is a field name and its context is a variable value
			err = filepath.Walk(filepath.Clean(variablesDir+"/"+variable.Name()), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					_, varFileName := filepath.Split(path)
					// Skip the symlink to a directory
					if strings.HasPrefix(varFileName, "..") {
						return filepath.SkipDir
					}
					varBytes, err := ioutil.ReadFile(path)
					if err != nil {
						return errors.Wrapf(err, "could not read variables variable %s", variable.Name())
					}

					// If variable type is password, set password value directly
					switch varFileName {
					case "password":
						staticVars[variable.Name()] = string(varBytes)
					default:
						staticVars[variable.Name()] = MergeStaticVar(staticVars[variable.Name()], varFileName, string(varBytes))
					}
				}
				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "could not read directory  %s", variable.Name())
			}

			vars = append(vars, staticVars)
		}
	}

	yamlBytes, err := InterpolateExplicitVariables(boshManifestBytes, vars, true)
	if err != nil {
		return errors.Wrap(err, "failed to interpolate explicit variables")
	}

	jsonBytes, err := json.Marshal(map[string]string{
		DesiredManifestKeyName: string(yamlBytes),
	})
	if err != nil {
		return errors.Wrapf(err, "could not marshal json output")
	}

	err = ioutil.WriteFile(outputFilePath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// InterpolateExplicitVariables interpolates explicit variables in the manifest
// Expects an array of maps, each element being a variable: [{ "name":"foo", "password": "value" }, {"name": "bar", "ca": "---"} ]
// Returns the new manifest as a byte array
func InterpolateExplicitVariables(boshManifestBytes []byte, vars []boshtpl.Variables, expectAllKeys bool) ([]byte, error) {
	multiVars := boshtpl.NewMultiVars(vars)
	tpl := boshtpl.NewTemplate(boshManifestBytes)

	// Following options are empty for cf-operator
	op := patch.Ops{}
	evalOpts := boshtpl.EvaluateOpts{
		ExpectAllKeys:     expectAllKeys,
		ExpectAllVarsUsed: false,
	}

	yamlBytes, err := tpl.Evaluate(multiVars, op, evalOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "could not evaluate variables")
	}

	m, err := LoadYAML(yamlBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "could not evaluate variables")
	}

	yamlBytes, err = m.Marshal()
	if err != nil {
		return nil, errors.Wrapf(err, "could not evaluate variables")
	}

	return yamlBytes, nil
}

// MergeStaticVar builds a map of values used for BOSH explicit variable interpolation
func MergeStaticVar(staticVar interface{}, field string, value string) interface{} {
	if staticVar == nil {
		staticVar = map[interface{}]interface{}{
			field: value,
		}
	} else {
		staticVarMap := staticVar.(map[interface{}]interface{})
		staticVarMap[field] = value
		staticVar = staticVarMap
	}

	return staticVar
}
