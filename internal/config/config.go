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
	return func(do *DeploymentOverrides) {
		do.Replicas = replicas
	}
}

func WithCPURequest(cpu string) Option {
	return func(do *DeploymentOverrides) {
		do.CPURequest = cpu
	}
}

func WithMemoryRequest(mem string) Option {
	return func(do *DeploymentOverrides) {
		do.MemoryRequest = mem
	}
}

func WithCPULimit(cpu string) Option {
	return func(do *DeploymentOverrides) {
		do.CPULimit = cpu
	}
}

func WithMemoryLimit(mem string) Option {
	return func(do *DeploymentOverrides) {
		do.MemoryLimit = mem
	}
}

func WithVolumeSize(size string) Option {
	return func(do *DeploymentOverrides) {
		do.VolumeSize = size
	}
}
