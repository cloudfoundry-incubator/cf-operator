package reference

import (
	"github.com/pkg/errors"

	bdv1 "code.cloudfoundry.org/quarks-operator/pkg/kube/apis/boshdeployment/v1alpha1"
)

// GetConfigMapsReferencedBy returns a list of all names for ConfigMaps referenced by the object
// The object can be an QuarksStatefulSet or a BOSHDeployment
func GetConfigMapsReferencedBy(object interface{}) (map[string]bool, error) {
	// Figure out the type of object
	switch object := object.(type) {
	case bdv1.BOSHDeployment:
		return getConfMapRefFromBdpl(object), nil
	default:
		return nil, errors.New("can't get config map references for unknown type; supported types are BOSHDeployment and QuarksStatefulSet")
	}
}

func getConfMapRefFromBdpl(object bdv1.BOSHDeployment) map[string]bool {
	result := map[string]bool{}

	if object.Spec.Manifest.Type == bdv1.ConfigMapReference {
		result[object.Spec.Manifest.Name] = true
	}

	for _, ops := range object.Spec.Ops {
		if ops.Type == bdv1.ConfigMapReference {
			result[ops.Name] = true
		}
	}

	return result
}
