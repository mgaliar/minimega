package rbac

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"phenix/api/config"
	v1 "phenix/types/version/v1"

	"github.com/mitchellh/mapstructure"
)

var NAME_TO_ROLE_CONFIG = map[string]string{
	"Global Admin":      "global-admin",
	"Global Viewer":     "global-viewer",
	"Experiment Admin":  "experiment-admin",
	"Experiment User":   "experiment-user",
	"Experiment Viewer": "experiment-viewer",
	"VM Viewer":         "vm-viewer",
}

type Role struct {
	Spec *v1.RoleSpec

	mappedPolicies map[string][]Policy
}

func RoleFromConfig(name string) (*Role, error) {
	if rname, ok := NAME_TO_ROLE_CONFIG[name]; ok {
		name = rname
	}

	c, err := config.Get("role/"+name, false)
	if err != nil {
		return nil, fmt.Errorf("getting role from store: %w", err)
	}

	var role v1.RoleSpec

	if err := mapstructure.Decode(c.Spec, &role); err != nil {
		return nil, fmt.Errorf("decoding role: %w", err)
	}

	return &Role{Spec: &role}, nil
}

func (this *Role) SetResourceNames(names ...string) error {
	// Gracefully handle when no names or a single empty name is passed,
	// defaulting to allow all.
	switch len(names) {
	case 0:
		names = []string{"*"}
	case 1:
		if names[0] == "" {
			names[0] = "*"
		}
	}

	for _, policy := range this.Spec.Policies {
		if policy.ResourceNames != nil {
			return fmt.Errorf("resource names already exist for policy")
		}

		for _, name := range names {
			// Checking to make sure pattern given in 'name' is valid. Thus, the
			// string provided to match it against is useless.
			if _, err := filepath.Match(name, "useless"); err != nil {
				continue
			}

			policy.ResourceNames = append(policy.ResourceNames, name)
		}
	}

	return nil
}

func (this *Role) AddPolicy(r, rn, v []string) {
	policy := &v1.PolicySpec{
		Resources:     r,
		ResourceNames: rn,
		Verbs:         v,
	}

	this.Spec.Policies = append(this.Spec.Policies, policy)
}

func (this Role) Allowed(resource, verb string, names ...string) bool {
	for _, policy := range this.policiesForResource(resource) {
		if policy.verbAllowed(verb) {
			if len(names) == 0 {
				return true
			}

			for _, n := range names {
				if policy.resourceNameAllowed(n) {
					return true
				}
			}
		}
	}

	return false
}

func (this Role) policiesForResource(resource string) []Policy {
	if err := this.mapPolicies(); err != nil {
		return nil
	}

	var policies []Policy

	for r, p := range this.mappedPolicies {
		if matched, _ := filepath.Match(r, resource); matched {
			policies = append(policies, p...)
			continue
		}
	}

	return policies
}

func (this *Role) mapPolicies() error {
	if this.mappedPolicies != nil {
		return nil
	}

	this.mappedPolicies = make(map[string][]Policy)

	var invalid []string

	for _, policy := range this.Spec.Policies {
		for _, resource := range policy.Resources {
			// Checking to make sure pattern given in 'resource' is valid. Thus, the
			// string provided to match it against is useless.
			if _, err := filepath.Match(resource, "useless"); err != nil {
				invalid = append(invalid, resource)
				continue
			}

			mapped := this.mappedPolicies[resource]
			mapped = append(mapped, Policy{Spec: policy})
			this.mappedPolicies[resource] = mapped
		}
	}

	if len(invalid) != 0 {
		return errors.New("invalid resource(s): " + strings.Join(invalid, ", "))
	}

	return nil
}
