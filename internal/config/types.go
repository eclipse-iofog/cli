package config

type Kind string

const (
	AgentConfigKind            Kind = "AgentConfig"
	CatalogItemKind            Kind = "CatalogItem"
	IofogctlConfigKind         Kind = "IofogctlConfig"
	IofogctlNamespaceKind      Kind = "IofogctlNamespace"
	RegistryKind               Kind = "Registry"
	VolumeKind                 Kind = "Volume"
	ControllerKind             Kind = "Controller"
	AgentKind                  Kind = "Agent"
	KubernetesControlPlaneKind Kind = "KubernetesControlPlane"
	RemoteControlPlaneKind     Kind = "RemoteControlPlane"
	LocalControlPlaneKind      Kind = "LocalControlPlane"
)

// Header contains k8s yaml header
type Header struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       Kind           `yaml:"kind" json:"kind"`
	Metadata   HeaderMetadata `yaml:"metadata" json:"metadata"`
	Spec       interface{}    `yaml:"spec" json:"spec"`
}

// Configuration contains the unmarshalled configuration file
type configuration struct {
	DefaultNamespace string `yaml:"defaultNamespace"`
}

type iofogctlConfig struct {
	Header `yaml:",inline"`
}

type iofogctlNamespace struct {
	Header `yaml:",inline"`
}

// HeaderMetadata contains k8s metadata
type HeaderMetadata struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}
