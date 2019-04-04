package manifest

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// DeploymentSecretType lists all the types of secrets used in
// the lifecycle of a BOSHDeployment
type DeploymentSecretType int

const (
	// DeploymentSecretTypeManifestWithOps is a manifest that has ops files applied
	DeploymentSecretTypeManifestWithOps DeploymentSecretType = iota
	// DeploymentSecretTypeManifestAndVars is a manifest whose variables have been interpolated
	DeploymentSecretTypeManifestAndVars
	// DeploymentSecretTypeGeneratedVariable is a BOSH variable generated using an ExtendedSecret
	DeploymentSecretTypeGeneratedVariable
	// DeploymentSecretTypeInstanceGroupResolvedProperties is a YAML file containing all properties needed to render an Instance Group
	DeploymentSecretTypeInstanceGroupResolvedProperties
)

func (s DeploymentSecretType) String() string {
	return [...]string{
		"with-ops",
		"with-vars",
		"var",
		"ig-resolved"}[s]
}

// CalculateSecretName generates a Secret name based on deployment name and variable name
func CalculateSecretName(secretType DeploymentSecretType, deployment string, variable string) string {
	if variable == "" {
		variable = secretType.String()
	} else {
		variable = fmt.Sprintf("%s-%s", secretType, variable)
	}

	nameRegex := regexp.MustCompile("[^-][a-z0-9-]*.[a-z0-9-]*[^-]")
	partRegex := regexp.MustCompile("[a-z0-9-]*")

	deploymentName := partRegex.FindString(strings.Replace(deployment, "_", "-", -1))
	variableName := partRegex.FindString(strings.Replace(variable, "_", "-", -1))
	secretName := nameRegex.FindString(deploymentName + "." + variableName)

	if len(secretName) > 63 {
		// secret names are limited to 63 characters so we recalculate the name as
		// <name trimmed to 31 characters><md5 hash of name>
		sumHex := md5.Sum([]byte(secretName))
		sum := hex.EncodeToString(sumHex[:])
		secretName = secretName[:63-32] + sum
	}

	return secretName
}

// CalculateEJobOutputSecretPrefixAndName generates a Secret prefix for the output
// of an Extended Job given a name, and calculates the final Secret name,
// given deployment name and container name
func CalculateEJobOutputSecretPrefixAndName(secretType DeploymentSecretType, deploymentName string, containerName string) (string, string) {
	prefix := CalculateSecretName(secretType, deploymentName, "")
	finalName := fmt.Sprintf("%s.%s", prefix, containerName)

	return prefix + ".", finalName
}
