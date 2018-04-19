package main

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
	"os"
)

type UsersSettings struct {
	Name      string         `json:"name"`
	Namespace string         `json:"namespace"`
	Tests     []TestSettings `json:"tests"`
}

type RoleSettings struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Rules     []Rule `json:"rules"`
}

type ClusterRoleSettings struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
	ApiGroups []string `json:"apiGroups"`
}

type ClusterRoleBindingSettings struct {
	Name     string            `json:"name"`
	Subjects []SubjectSettings `json:"subjects"`
	Role     RoleRefSettings   `json:"role"`
}

type RoleBindingSettings struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Subjects  []SubjectSettings `json:"subjects"`
	Role      RoleRefSettings   `json:"role"`
}

type RoleRefSettings struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	ApiGroup string `json:"apiGroup"`
}

type SubjectSettings struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	ApiGroup  string `json:"apiGroup"`
	Namespace string `json:"namespace"`
}

type BootstrapSettings struct {
	Users               []UsersSettings              `json:"serviceAccounts"`
	Namespaces          []string                     `json:"namespaces"`
	Roles               []RoleSettings               `json:"roles"`
	ClusterRoles        []ClusterRoleSettings        `json:"clusterRoles"`
	RoleBindings        []RoleBindingSettings        `json:"roleBindings"`
	ClusterRoleBindings []ClusterRoleBindingSettings `json:"clusterRoleBindings"`
}

type TestSettings struct {
	Path         string `json:"path"`
	ExpectedCode int    `json:"expectedCode"`
}

// Load settings from bootstrap.json
func LoadBootstrapSettings(path string) (*BootstrapSettings, error) {
	conf := config.NewConfig()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	source := file.NewSource(file.WithPath(path))
	err := conf.Load(source)
	if err != nil {
		return nil, err
	}
	var settings = new(BootstrapSettings)
	if err = conf.Get().Scan(settings); err != nil {
		return nil, err
	}
	return settings, nil
}
