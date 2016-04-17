package config

import (
	"github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/docker/libcompose/utils"
)

// MergeServicesV2 merges a compose file into an existing set of service configs
func MergeServicesV2(existingServices *Configs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte) (map[string]*ServiceConfig, error) {
	configs := make(map[string]*ServiceConfig)

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	datas := config.Services

	if err := Interpolate(environmentLookup, &datas); err != nil {
		return nil, err
	}

	// TODO: v2 validation
	/*if err := validate(datas); err != nil {
		return nil, err
	}*/

	for name, data := range datas {
		data, err := parse(resourceLookup, environmentLookup, file, data, datas)
		if err != nil {
			logrus.Errorf("Failed to parse service %s: %v", name, err)
			return nil, err
		}

		if serviceConfig, ok := existingServices.Get(name); ok {
			var rawExistingService RawService
			if err := utils.Convert(serviceConfig, &rawExistingService); err != nil {
				return nil, err
			}

			data = mergeConfig(rawExistingService, data)
		}

		datas[name] = data
	}

	// TODO: v2 validation
	/*for name, data := range datas {
		err := validateServiceConstraints(data, name)
		if err != nil {
			return nil, err
		}
	}*/

	if err := utils.Convert(datas, &configs); err != nil {
		return nil, err
	}

	// TODO
	//adjustValues(configs)

	return configs, nil
}

// ParseVolumes
func ParseVolumes(environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte) (map[string]*VolumeConfig, error) {
	volumeConfigs := make(map[string]*VolumeConfig)

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	if err := utils.Convert(config.Volumes, &volumeConfigs); err != nil {
		return nil, err
	}

	return volumeConfigs, nil
}

// ParseNetworks
func ParseNetworks(environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte) (map[string]*NetworkConfig, error) {
	networkConfigs := make(map[string]*NetworkConfig)

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	if err := utils.Convert(config.Networks, &networkConfigs); err != nil {
		return nil, err
	}

	return networkConfigs, nil
}
