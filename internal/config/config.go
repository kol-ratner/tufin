package config

type DeploymentOverrides struct {
	Replicas      int32
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
	VolumeSize    string
}

type Option func(*DeploymentOverrides)

func WithReplicas(replicas int32) Option {
	return func(co *DeploymentOverrides) {
		co.Replicas = replicas
	}
}

func WithCPURequest(cpu string) Option {
	return func(co *DeploymentOverrides) {
		co.CPURequest = cpu
	}
}

func WithMemoryRequest(mem string) Option {
	return func(co *DeploymentOverrides) {
		co.MemoryRequest = mem
	}
}

func WithCPULimit(cpu string) Option {
	return func(co *DeploymentOverrides) {
		co.CPULimit = cpu
	}
}

func WithMemoryLimit(mem string) Option {
	return func(co *DeploymentOverrides) {
		co.MemoryLimit = mem
	}
}

func WithVolumeSize(size string) Option {
	return func(co *DeploymentOverrides) {
		co.VolumeSize = size
	}
}
