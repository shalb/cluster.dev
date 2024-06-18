package types

type ExecNestedSchema struct {
	APIVersion string            `yaml:"api_version,omitempty" json:"api_version,omitempty"`
	Args       []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Command    string            `yaml:"command,omitempty" json:"command,omitempty"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type ProviderConfigSpec struct {
	ConfigPath            string            `yaml:"config_path,omitempty" json:"config_path,omitempty"`
	ConfigPaths           []string          `yaml:"config_paths,omitempty" json:"config_paths,omitempty"`
	ClientCertificate     string            `yaml:"client_certificate,omitempty" json:"client_certificate,omitempty"`
	ClientKey             string            `yaml:"client_key,omitempty" json:"client_key,omitempty"`
	ClusterCaCertificate  string            `yaml:"cluster_ca_certificate,omitempty" json:"cluster_ca_certificate,omitempty"`
	ConfigContext         string            `yaml:"config_context,omitempty" json:"config_context,omitempty"`
	ConfigContextCluster  string            `yaml:"config_context_cluster,omitempty" json:"config_context_cluster,omitempty"`
	ConfigContextUser     string            `yaml:"config_context_user,omitempty"  json:"config_context_user,omitempty"`
	ConfigContextAuthInfo string            `yaml:"config_context_auth_info,omitempty"  json:"config_context_auth_info,omitempty"`
	ProxyURL              string            `yaml:"proxy_url,omitempty"  json:"proxy_url,omitempty"`
	Exec                  *ExecNestedSchema `yaml:"exec,omitempty" json:"exec,omitempty"`
	Host                  string            `yaml:"host,omitempty" json:"host,omitempty"`
	Insecure              string            `yaml:"insecure,omitempty" json:"insecure,omitempty"`
	Password              string            `yaml:"password,omitempty" json:"password,omitempty"`
	Token                 string            `yaml:"token,omitempty" json:"token,omitempty"`
	Username              string            `yaml:"username,omitempty" json:"username,omitempty"`
	IgnoreAnnotations     string            `yaml:"ignore_annotations,omitempty" json:"ignore_annotations,omitempty"`
	IgnoreLabels          string            `yaml:"ignore_labels,omitempty" json:"ignore_labels,omitempty"`
	TlsServerName         string            `yaml:"tls_server_name,omitempty" json:"tls_server_name,omitempty"`
}
