package bpm

import (
	"sort"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// Hooks from a BPM config
type Hooks struct {
	PreStart string `yaml:"pre_start,omitempty" json:"pre_start,omitempty"`
}

// Limits from a BPM config
type Limits struct {
	Memory    string `yaml:"memory,omitempty" json:"memory,omitempty"`
	OpenFiles int    `yaml:"open_files,omitempty" json:"open_files,omitempty"`
	Processes int    `yaml:"processes,omitempty" json:"processes,omitempty"`
}

// Volume from a BPM config
type Volume struct {
	Path            string `yaml:"path,omitempty" json:"path,omitempty"`
	Writable        bool   `yaml:"writable,omitempty" json:"writable,omitempty"`
	AllowExecutions bool   `yaml:"allow_executions,omitempty" json:"allow_executions,omitempty"`
	MountOnly       bool   `yaml:"mount_only,omitempty" json:"mount_only,omitempty"`
}

// Unsafe from a BPM config
type Unsafe struct {
	Privileged          bool     `yaml:"privileged,omitempty" json:"privileged,omitempty"`
	UnrestrictedVolumes []Volume `yaml:"unrestricted_volumes,omitempty" json:"unrestricted_volumes,omitempty"`
}

// Process from a BPM config
type Process struct {
	Name              string              `yaml:"name,omitempty" json:"name,omitempty"`
	Executable        string              `yaml:"executable,omitempty" json:"executable,omitempty"`
	Args              []string            `yaml:"args,omitempty" json:"args,omitempty"`
	Env               map[string]string   `yaml:"env,omitempty" json:"env,omitempty"`
	Workdir           string              `yaml:"workdir,omitempty" json:"workdir,omitempty"`
	Hooks             Hooks               `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Capabilities      []string            `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Limits            Limits              `yaml:"limits,omitempty" json:"limits,omitempty"`
	Requests          corev1.ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList,castkey=ResourceName"`
	EphemeralDisk     bool                `yaml:"ephemeral_disk,omitempty" json:"ephemeral_disk,omitempty"`
	PersistentDisk    bool                `yaml:"persistent_disk,omitempty" json:"persistent_disk,omitempty"`
	AdditionalVolumes []Volume            `yaml:"additional_volumes,omitempty" json:"additional_volumes,omitempty"`
	Unsafe            Unsafe              `yaml:"unsafe,omitempty" json:"unsafe,omitempty"`
}

// Config represent a BPM configuration
type Config struct {
	Processes           []Process `yaml:"processes,omitempty" json:"processes,omitempty"`
	UnsupportedTemplate bool      `json:"unsupported_template"`
}

// Configs holds a collection of BPM configurations by their according job
type Configs map[string]Config

// NewConfig creates a new Config object from the yaml
func NewConfig(data []byte) (Config, error) {
	config := Config{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, errors.Wrapf(err, "Unmarshalling data %s failed", string(data))
	}
	return config, nil
}

// NewEnvs returns a list of k8s env vars, based on the bpm envs, overwritten
// by the overrides list passed to the function.
func (p *Process) NewEnvs(overrides []corev1.EnvVar) []corev1.EnvVar {
	seen := make(map[string]corev1.EnvVar)

	for name, value := range p.Env {
		seen[name] = corev1.EnvVar{Name: name, Value: value}
	}

	for _, env := range overrides {
		seen[env.Name] = env
	}

	result := make([]corev1.EnvVar, 0, len(seen))
	for _, value := range seen {
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// UpdateEnv adds the overrides env vars to the env list of the bpm process
func (p *Process) UpdateEnv(overrides []corev1.EnvVar) {
	if p.Env == nil {
		p.Env = map[string]string{}
	}
	for _, env := range overrides {
		if env.Value == "" && env.ValueFrom != nil {
			p.Env[env.Name] = env.ValueFrom.String()
		} else {
			p.Env[env.Name] = env.Value
		}
	}
}
