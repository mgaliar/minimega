package config

import (
	"errors"
	"fmt"
	"strings"

	"phenix/store"
	"phenix/types"
	"phenix/types/version"
	"phenix/util"
	"phenix/util/editor"

	"gopkg.in/yaml.v3"
)

// ConfigHook is a function to be called during the different lifecycle stages
// of a config. The passed config can be updated by the hook functions as
// necessary, and an error can be returned if the lifecycle stage should be
// halted.
type ConfigHook func(string, *types.Config) error

var hooks = make(map[string][]ConfigHook)

// RegisterConfigHook registers a ConfigHook for the given config kind.
func RegisterConfigHook(kind string, hook ConfigHook) {
	hooks[kind] = append(hooks[kind], hook)
}

func Init() error {
	for _, file := range AssetNames() {
		var c types.Config

		if err := yaml.Unmarshal(MustAsset(file), &c); err != nil {
			return fmt.Errorf("unmarshaling default config %s: %w", file, err)
		}

		name := strings.ToLower(c.Kind) + "/" + c.Metadata.Name

		// Don't attempt to create this default config again if it already exists in
		// the store.
		if _, err := Get(name, false); err == nil {
			continue
		}

		if err := store.Create(&c); err != nil {
			return fmt.Errorf("storing default config %s: %w", file, err)
		}
	}

	return nil
}

// List collects configs of the given type (topology, scenario, experiment). If
// no config type is specified, or `all` is specified, then all the known
// configs will be collected. It returns a slice of configs and any errors
// encountered while getting the configs from the store.
func List(which string) (types.Configs, error) {
	var (
		configs types.Configs
		err     error
	)

	switch which {
	case "", "all":
		configs, err = store.List("Topology", "Scenario", "Experiment", "Image", "User", "Role")
	case "topology":
		configs, err = store.List("Topology")
	case "scenario":
		configs, err = store.List("Scenario")
	case "experiment":
		configs, err = store.List("Experiment")
	case "image":
		configs, err = store.List("Image")
	case "user":
		configs, err = store.List("User")
	case "role":
		configs, err = store.List("Role")
	default:
		return nil, util.HumanizeError(fmt.Errorf("unknown config kind provided: %s", which), "")
	}

	if err != nil {
		return nil, fmt.Errorf("getting list of configs from store: %w", err)
	}

	return configs, nil
}

// Get retrieves the config with the given name. The given name should be of the
// form `type/name`, where `type` is one of `topology, scenario, or experiment`.
// It returns a pointer to the config and any errors encountered while getting
// the config from the store. Note that the returned config will **not** have
// its `spec` and `status` fields casted to the given type, but instead will be
// generic `map[string]interface{}` fields. It's up to the caller to convert
// these fields into the appropriate types.
func Get(name string, upgrade bool) (*types.Config, error) {
	if name == "" {
		return nil, util.HumanizeError(fmt.Errorf("no config name provided"), "")
	}

	c, err := types.NewConfig(name)
	if err != nil {
		return nil, err
	}

	if err := store.Get(c); err != nil {
		return nil, fmt.Errorf("getting config from store: %w", err)
	}

	if upgrade {
		latest := version.StoredVersion[c.Kind]

		if c.APIVersion() != latest {
			upgrader := types.GetUpgrader(c.Kind + "/" + latest)
			if upgrader != nil {
				iface, err := upgrader.Upgrade(c.APIVersion(), c.Spec, c.Metadata)
				if err != nil {
					return nil, fmt.Errorf("upgrading config: %w", err)
				}

				cfg, err := types.NewConfigFromSpec(c.Metadata.Name, iface)
				if err != nil {
					return nil, fmt.Errorf("creating new config from spec: %w", err)
				}

				return cfg, nil
			}
		}
	}

	return c, nil
}

// Create reads a config file from the given path, validates it, and persists it
// to the store. Validation of configs is done against OpenAPIv3 schema
// definitions. In the event the config file being read defines an experiment,
// additional validations are done to ensure the annotated topology (required)
// and scenario (optional) exist. It returns a pointer to the resulting config
// struct and eny errors encountered while creating the config.
func Create(path string, validate bool) (*types.Config, error) {
	if path == "" {
		return nil, fmt.Errorf("no config file provided")
	}

	c, err := types.NewConfigFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("creating new config from file: %w", err)
	}

	if validate {
		if err := types.ValidateConfigSpec(*c); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}
	}

	for _, hook := range hooks[c.Kind] {
		if err := hook("create", c); err != nil {
			return nil, fmt.Errorf("calling config hook: %w", err)
		}

		if validate {
			// Validate again since config hooks can modify the config.
			if err := types.ValidateConfigSpec(*c); err != nil {
				return nil, fmt.Errorf("validating config after config hook: %w", err)
			}
		}
	}

	if err := store.Create(c); err != nil {
		return nil, fmt.Errorf("storing config: %w", err)
	}

	return c, nil
}

// Edit retrieves the config with the given name for editing. The given name
// should be of the form `type/name`, where `type` is one of `topology,
// scenario, or experiment`. A YAML representation of the config is written to a
// temporary file, and that file is opened for editing using the default editor
// (as defined by the user's `EDITOR` env variable). If no default editor is
// found, `vim` is used. If no changes were made to the file, an error of type
// `editor.ErrNoChange` is returned. This can be checked using the
// `IsConfigNotModified` function. It returns the updated config and any errors
// encountered while editing the config.
func Edit(name string, force bool) (*types.Config, error) {
	if name == "" {
		return nil, fmt.Errorf("no config name provided")
	}

	c, err := types.NewConfig(name)
	if err != nil {
		return nil, err
	}

	if err := store.Get(c); err != nil {
		return nil, fmt.Errorf("getting config from store: %w", err)
	}

	if !force && c.Kind == "Experiment" {
		exp, err := types.DecodeExperimentFromConfig(*c)
		if err != nil {
			return nil, fmt.Errorf("decoding experiment from config: %w", err)
		}

		if exp.Running() {
			return nil, fmt.Errorf("cannot edit running experiment")
		}
	}

	body, err := yaml.Marshal(c.Spec)
	if err != nil {
		return nil, fmt.Errorf("marshaling config to YAML: %w", err)
	}

	body, err = editor.EditData(body)
	if err != nil {
		return nil, fmt.Errorf("editing config: %w", err)
	}

	var spec map[string]interface{}

	if err := yaml.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("unmarshaling config as YAML: %w", err)
	}

	c.Spec = spec

	// TODO: validate after edit

	for _, hook := range hooks[c.Kind] {
		if err := hook("edit", c); err != nil {
			return nil, fmt.Errorf("calling config hook: %w", err)
		}
	}

	if err := store.Update(c); err != nil {
		return nil, fmt.Errorf("updating config in store: %w", err)
	}

	return c, nil
}

// Delete removes the config with the given name from the store. The given name
// should be of the form `type/name`, where `type` is one of `topology,
// scenario, or experiment`. If `all` is specified, then all the known configs
// are removed. It returns any errors encountered while removing the config from
// the store.
func Delete(name string) error {
	if name == "" {
		return fmt.Errorf("no config name provided")
	}

	if name == "all" {
		configs, _ := List("all")

		for _, c := range configs {
			if err := delete(&c); err != nil {
				return err
			}
		}

		return nil
	}

	c, err := Get(name, false)
	if err != nil {
		return fmt.Errorf("getting config '%s': %w", name, err)
	}

	return delete(c)
}

func delete(c *types.Config) error {
	if err := store.Delete(c); err != nil {
		return fmt.Errorf("deleting config in store: %w", err)
	}

	return nil
}

// IsConfigNotModified returns a boolean indicating whether the error is known
// to report that a config was not modified during editing. It is satisfied by
// editor.ErrNoChange.
func IsConfigNotModified(err error) bool {
	return errors.Is(err, editor.ErrNoChange)
}
