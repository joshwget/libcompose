package service

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/yaml"
	shlex "github.com/flynn/go-shlex"
	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	exp := []string{"sh", "-c", "exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}"}
	cmd, err := shlex.Split("sh -c 'exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}'")
	assert.Nil(t, err)
	assert.Equal(t, exp, cmd)
}

func TestParseBindsAndVolumes(t *testing.T) {
	ctx := &ctx.Context{}
	ctx.ComposeFiles = []string{"foo/docker-compose.yml"}
	ctx.ResourceLookup = &lookup.FileResourceLookup{}

	abs, err := filepath.Abs(".")
	assert.Nil(t, err)
	cfg, hostCfg, err := Convert(&config.ServiceConfig{
		Volumes: &yaml.Volumes{
			Volumes: []*yaml.Volume{
				{
					Destination: "/foo",
				},
				{
					Source:      "/home",
					Destination: "/home",
				},
				{
					Destination: "/bar/baz",
				},
				{
					Source:      ".",
					Destination: "/home",
				},
				{
					Source:      "/usr/lib",
					Destination: "/usr/lib",
					AccessMode:  "ro",
				},
			},
		},
	}, ctx.Context, nil)
	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{"/foo": {}, "/bar/baz": {}}, cfg.Volumes)
	assert.Equal(t, []string{"/home:/home", abs + "/foo:/home", "/usr/lib:/usr/lib:ro"}, hostCfg.Binds)
}

func TestParseLabels(t *testing.T) {
	ctx := &ctx.Context{}
	ctx.ComposeFiles = []string{"foo/docker-compose.yml"}
	ctx.ResourceLookup = &lookup.FileResourceLookup{}
	bashCmd := "bash"
	fooLabel := "foo.label"
	fooLabelValue := "service.config.value"
	sc := &config.ServiceConfig{
		Entrypoint: yaml.Command([]string{bashCmd}),
		Labels:     yaml.SliceorMap{fooLabel: "service.config.value"},
	}
	cfg, _, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	cfg.Labels[fooLabel] = "FUN"
	cfg.Entrypoint[0] = "less"

	assert.Equal(t, fooLabelValue, sc.Labels[fooLabel])
	assert.Equal(t, "FUN", cfg.Labels[fooLabel])

	assert.Equal(t, yaml.Command{bashCmd}, sc.Entrypoint)
	assert.Equal(t, []string{"less"}, []string(cfg.Entrypoint))
}

func TestBlkioWeight(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		BlkioWeight: 10,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, uint16(10), hostCfg.BlkioWeight)
}

func TestBlkioWeightDevices(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		BlkioWeightDevice: []string{
			"/dev/sda:10",
		},
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]*blkiodev.WeightDevice{
		&blkiodev.WeightDevice{
			Path:   "/dev/sda",
			Weight: 10,
		},
	}, hostCfg.BlkioWeightDevice))
}

func TestCPUPeriod(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		CPUPeriod: 50000,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, int64(50000), hostCfg.CPUPeriod)
}

func TestDNSOpts(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DNSOpts: []string{
			"use-vc",
			"no-tld-query",
		},
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]string{
		"use-vc",
		"no-tld-query",
	}, hostCfg.DNSOptions))
}

func TestMemReservation(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		MemReservation: 100000,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, int64(100000), hostCfg.MemoryReservation)
}

func TestIsolation(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		Isolation: "default",
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, container.Isolation("default"), hostCfg.Isolation)
}

func TestOomKillDisable(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		OomKillDisable: true,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, true, *hostCfg.OomKillDisable)
}

func TestBlkioDeviceReadBps(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DeviceReadBps: yaml.MaporColonSlice([]string{
			"/dev/sda:100000",
		}),
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]*blkiodev.ThrottleDevice{
		&blkiodev.ThrottleDevice{
			Path: "/dev/sda",
			Rate: 100000,
		},
	}, hostCfg.BlkioDeviceReadBps))
}

func TestBlkioDeviceReadIOps(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DeviceReadIOps: yaml.MaporColonSlice([]string{
			"/dev/sda:100000",
		}),
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]*blkiodev.ThrottleDevice{
		&blkiodev.ThrottleDevice{
			Path: "/dev/sda",
			Rate: 100000,
		},
	}, hostCfg.BlkioDeviceReadIOps))
}

func TestBlkioDeviceWriteBps(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DeviceWriteBps: yaml.MaporColonSlice([]string{
			"/dev/sda:100000",
		}),
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]*blkiodev.ThrottleDevice{
		&blkiodev.ThrottleDevice{
			Path: "/dev/sda",
			Rate: 100000,
		},
	}, hostCfg.BlkioDeviceWriteBps))
}

func TestBlkioDeviceWriteIOps(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DeviceWriteIOps: yaml.MaporColonSlice([]string{
			"/dev/sda:100000",
		}),
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]*blkiodev.ThrottleDevice{
		&blkiodev.ThrottleDevice{
			Path: "/dev/sda",
			Rate: 100000,
		},
	}, hostCfg.BlkioDeviceWriteIOps))
}
