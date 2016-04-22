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
		data, err := parseV2(resourceLookup, environmentLookup, file, data, datas)
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

func parseV2(resourceLookup ResourceLookup, environmentLookup EnvironmentLookup, inFile string, serviceData RawService, datas RawServiceMap) (RawService, error) {
	serviceData, err := readEnvFileV2(resourceLookup, inFile, serviceData)
	if err != nil {
		return nil, err
	}

	serviceData, err = resolveBuild(inFile, serviceData)
	if err != nil {
		return nil, err
	}

	value, ok := serviceData["extends"]
	if !ok {
		return serviceData, nil
	}

	mapValue, ok := value.(map[interface{}]interface{})
	if !ok {
		return serviceData, nil
	}

	if resourceLookup == nil {
		return nil, fmt.Errorf("Can not use extends in file %s no mechanism provided to files", inFile)
	}

	file := asString(mapValue["file"])
	service := asString(mapValue["service"])

	if service == "" {
		return serviceData, nil
	}

	var baseService RawService

	if file == "" {
		if serviceData, ok := datas[service]; ok {
			baseService, err = parseV2(resourceLookup, environmentLookup, inFile, serviceData, datas)
		} else {
			return nil, fmt.Errorf("Failed to find service %s to extend", service)
		}
	} else {
		bytes, resolved, err := resourceLookup.Lookup(file, inFile)
		if err != nil {
			logrus.Errorf("Failed to lookup file %s: %v", file, err)
			return nil, err
		}

		var baseRawServices RawServiceMap
		if err := yaml.Unmarshal(bytes, &baseRawServices); err != nil {
			return nil, err
		}

		err = Interpolate(environmentLookup, &baseRawServices)
		if err != nil {
			return nil, err
		}

		if err := validate(baseRawServices); err != nil {
			return nil, err
		}

		baseService, ok = baseRawServices[service]
		if !ok {
			return nil, fmt.Errorf("Failed to find service %s in file %s", service, file)
		}

		baseService, err = parseV2(resourceLookup, environmentLookup, resolved, baseService, baseRawServices)
	}

	if err != nil {
		return nil, err
	}

	baseService = clone(baseService)

	logrus.Debugf("Merging %#v, %#v", baseService, serviceData)

	for _, k := range noMerge {
		if _, ok := baseService[k]; ok {
			source := file
			if source == "" {
				source = inFile
			}
			return nil, fmt.Errorf("Cannot extend service '%s' in %s: services with '%s' cannot be extended", service, source, k)
		}
	}

	baseService = mergeConfig(baseService, serviceData)

	logrus.Debugf("Merged result %#v", baseService)

	return baseService, nil
}
