package config

import yaml "github.com/cloudfoundry-incubator/candiedyaml"

var (
	// ValidRemotes list the of valid prefixes that can be sent to Docker as a build remote location
	// This is public for consumers of libcompose to use
	ValidRemotes = []string{
		"git://",
		"git@github.com:",
		"github.com",
		"http:",
		"https:",
	}
	noMerge = []string{
		"links",
		"volumes_from",
	}
)

// MergeServices merges a compose file into an existing set of service configs
func Merge(existingServices *Configs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte) (map[string]*ServiceConfig, map[string]*VolumeConfig, map[string]*NetworkConfig, error) {
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, nil, nil, err
	}

	if config.Version == "2" {
		services, err := MergeServicesV2(existingServices, environmentLookup, resourceLookup, file, bytes)
		if err != nil {
			return nil, nil, nil, err
		}
		volumes, err := ParseVolumes(environmentLookup, resourceLookup, file, bytes)
		if err != nil {
			return nil, nil, nil, err
		}
		networks, err := ParseNetworks(environmentLookup, resourceLookup, file, bytes)
		if err != nil {
			return nil, nil, nil, err
		}
		return services, volumes, networks, nil
	} else {
		v1Services, err := MergeServicesV1(existingServices, environmentLookup, resourceLookup, file, bytes)
		if err != nil {
			return nil, nil, nil, err
		}
		return ConvertV1toV2(v1Services, environmentLookup, resourceLookup)
	}
}
