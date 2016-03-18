package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "gopkg.in/check.v1"
	"path/filepath"
)

func (s *RunSuite) TestFields(c *C) {
	p := s.CreateProjectFromText(c, `
        hello:
          image: tianon/true
          cpuset: 0,1
          mem_limit: 4194304
        `)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.Config.Image, Equals, "tianon/true")
	c.Assert(cn.HostConfig.CpusetCpus, Equals, "0,1")
	c.Assert(cn.HostConfig.Memory, Equals, int64(4194304))
}

func (s *RunSuite) TestEmptyEntrypoint(c *C) {
	p := s.CreateProjectFromText(c, `
        nil-cmd:
          image: busybox
          entrypoint: []
        `)

	name := fmt.Sprintf("%s_%s_1", p, "nil-cmd")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.Config.Entrypoint, IsNil)
}

func (s *RunSuite) TestHelloWorld(c *C) {
	p := s.CreateProjectFromText(c, `
        hello:
          image: tianon/true
        `)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.Name, Equals, "/"+name)
}

func (s *RunSuite) TestContainerName(c *C) {
	containerName := "containerName"
	template := fmt.Sprintf(`hello:
    image: busybox
    command: top
    container_name: %s`, containerName)
	s.CreateProjectFromText(c, template)

	cn := s.GetContainerByName(c, containerName)
	c.Assert(cn, NotNil)

	c.Assert(cn.Name, Equals, "/"+containerName)
}

func (s *RunSuite) TestContainerNameWithScale(c *C) {
	containerName := "containerName"
	template := fmt.Sprintf(`hello:
    image: busybox
    command: top
    container_name: %s`, containerName)
	p := s.CreateProjectFromText(c, template)

	s.FromText(c, p, "scale", "hello=2", template)
	containers := s.GetContainersByProject(c, p)
	c.Assert(len(containers), Equals, 1)

}

func (s *RunSuite) TestInterpolation(c *C) {
	os.Setenv("IMAGE", "tianon/true")

	p := s.CreateProjectFromText(c, `
        test:
          image: $IMAGE
        `)

	name := fmt.Sprintf("%s_%s_1", p, "test")
	testContainer := s.GetContainerByName(c, name)

	p = s.CreateProjectFromText(c, `
        reference:
          image: tianon/true
        `)

	name = fmt.Sprintf("%s_%s_1", p, "reference")
	referenceContainer := s.GetContainerByName(c, name)

	c.Assert(testContainer, NotNil)

	c.Assert(referenceContainer.Image, Equals, testContainer.Image)

	os.Unsetenv("IMAGE")
}

func (s *RunSuite) TestInterpolationWithExtends(c *C) {
	os.Setenv("IMAGE", "tianon/true")
	os.Setenv("TEST_PORT", "8000")

	p := s.CreateProjectFromText(c, `
        test:
                extends:
                        file: ./assets/interpolation/docker-compose.yml
                        service: base
                ports:
                        - ${TEST_PORT}
        `)

	name := fmt.Sprintf("%s_%s_1", p, "test")
	testContainer := s.GetContainerByName(c, name)

	p = s.CreateProjectFromText(c, `
	reference:
	  image: tianon/true
		ports:
		  - 8000
	`)

	name = fmt.Sprintf("%s_%s_1", p, "reference")
	referenceContainer := s.GetContainerByName(c, name)

	c.Assert(testContainer, NotNil)

	c.Assert(referenceContainer.Image, Equals, testContainer.Image)

	os.Unsetenv("TEST_PORT")
	os.Unsetenv("IMAGE")
}

func (s *RunSuite) TestFieldTypeConversions(c *C) {
	os.Setenv("LIMIT", "40000000")

	p := s.CreateProjectFromText(c, `
        test:
          image: tianon/true
          mem_limit: $LIMIT
          memswap_limit: "40000000"
        `)

	name := fmt.Sprintf("%s_%s_1", p, "test")
	testContainer := s.GetContainerByName(c, name)

	p = s.CreateProjectFromText(c, `
        reference:
          image: tianon/true
          mem_limit: 40000000
          memswap_limit: 40000000
        `)

	name = fmt.Sprintf("%s_%s_1", p, "reference")
	referenceContainer := s.GetContainerByName(c, name)

	c.Assert(testContainer, NotNil)

	c.Assert(referenceContainer.Image, Equals, testContainer.Image)

	os.Unsetenv("LIMIT")
}

func (s *RunSuite) TestMultipleComposeFilesOneTwo(c *C) {
	p := "multiple"
	cmd := exec.Command(s.command, "-f", "./assets/multiple/one.yml", "-f", "./assets/multiple/two.yml", "create")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	c.Assert(err, IsNil)

	containerNames := []string{"multiple", "simple", "another", "yetanother"}

	for _, containerName := range containerNames {
		name := fmt.Sprintf("%s_%s_1", p, containerName)
		container := s.GetContainerByName(c, name)

		c.Assert(container, NotNil)
	}

	name := fmt.Sprintf("%s_%s_1", p, "multiple")
	container := s.GetContainerByName(c, name)

	c.Assert(container.Config.Image, Equals, "busybox")
	c.Assert(container.Config.Cmd.Slice(), DeepEquals, []string{"echo", "two"})
	c.Assert(container.Config.Env, DeepEquals, []string{"KEY1=VAL1", "KEY2=VAL2"})
}

func (s *RunSuite) TestMultipleComposeFilesTwoOne(c *C) {
	p := "multiple"
	cmd := exec.Command(s.command, "-f", "./assets/multiple/two.yml", "-f", "./assets/multiple/one.yml", "create")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	c.Assert(err, IsNil)

	containerNames := []string{"multiple", "simple", "another", "yetanother"}

	for _, containerName := range containerNames {
		name := fmt.Sprintf("%s_%s_1", p, containerName)
		container := s.GetContainerByName(c, name)

		c.Assert(container, NotNil)
	}

	name := fmt.Sprintf("%s_%s_1", p, "multiple")
	container := s.GetContainerByName(c, name)

	c.Assert(container.Config.Image, Equals, "tianon/true")
	c.Assert(container.Config.Cmd.Slice(), DeepEquals, []string{"echo", "two"})
	c.Assert(container.Config.Env, DeepEquals, []string{"KEY2=VAL2", "KEY1=VAL1"})
}

func (s *RunSuite) TestDefaultMultipleComposeFiles(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(filepath.Join("../../", s.command), "-p", p, "create")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./assets/multiple-composefiles-default/"
	err := cmd.Run()

	c.Assert(err, IsNil)

	containerNames := []string{"simple", "another", "yetanother"}

	for _, containerName := range containerNames {
		name := fmt.Sprintf("%s_%s_1", p, containerName)
		container := s.GetContainerByName(c, name)

		c.Assert(container, NotNil)
	}
}

func (s *RunSuite) TestValidation(c *C) {
	template := `
  test:
    image: busybox
    ports: invalid_type
	`
	_, output := s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'test' configuration key 'ports' contains an invalid type, it should be an array."), Equals, true)

	template = `
  test:
    image: busybox
    build: .
	`
	_, output = s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'test' has both an image and build path specified. A service can either be built to image or use an existing image, not both."), Equals, true)

	template = `
  test:
    image: busybox
    ports: invalid_type
    links: invalid_type
    devices:
      - /dev/foo:/dev/foo
      - /dev/foo:/dev/foo
  `
	_, output = s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'test' configuration key 'ports' contains an invalid type, it should be an array."), Equals, true)
	c.Assert(strings.Contains(output, "Service 'test' configuration key 'links' contains an invalid type, it should be an array"), Equals, true)
	c.Assert(strings.Contains(output, "Service 'test' configuration key 'devices' value [/dev/foo:/dev/foo /dev/foo:/dev/foo] has non-unique elements"), Equals, true)
}

func (s *RunSuite) TestValidationWithExtends(c *C) {
	template := `
  base:
    image: busybox
    privilege: "something"
  test:
    extends:
      service: base
	`

	_, output := s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Unsupported config option for base service: 'privilege' (did you mean 'privileged'?)"), Equals, true)

	template = `
  base:
    image: busybox
  test:
    extends:
      service: base
    links: invalid_type
	`

	_, output = s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'test' configuration key 'links' contains an invalid type, it should be an array"), Equals, true)

	template = `
  test:
    extends:
      file: ./assets/validation/valid/docker-compose.yml
      service: base
    devices:
      - /dev/foo:/dev/foo
      - /dev/foo:/dev/foo
	`

	_, output = s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'test' configuration key 'devices' value [/dev/foo:/dev/foo /dev/foo:/dev/foo] has non-unique elements"), Equals, true)

	template = `
  test:
    extends:
      file: ./assets/validation/invalid/docker-compose.yml
      service: base
	`

	_, output = s.FromTextCaptureOutput(c, s.RandomProject(), "create", template)

	c.Assert(strings.Contains(output, "Service 'base' configuration key 'ports' contains an invalid type, it should be an array."), Equals, true)
}
