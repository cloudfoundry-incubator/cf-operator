package manifest

import (
	"code.cloudfoundry.org/cf-operator/pkg/bosh/bpm"
)

// BPMInfo contains custome information about
// instance group which matters for exsts pods
// such as AZ's, instance group count and BPM Configs
type BPMInfo struct {
	InstanceGroup BPMInstanceGroup `json:"instance_group,omitempty"`
	Configs       bpm.Configs      `json:"configs,omitempty"`
	Variables     []Variable       `json:"variables,omitempty"`
}

// BPMInstanceGroup is a custome instance group spec
// that should be included in the BPM secret created
// by the bpm extendedjob.
type BPMInstanceGroup struct {
	Name      string   `json:"name"`
	Instances int      `json:"instances"`
	AZs       []string `json:"azs"`
	Env       AgentEnv `json:"env,omitempty"`
}
