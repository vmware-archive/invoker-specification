package framework

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Suite struct {
	Name        string
	Description string
	Cases       []*Testcase
	Port        int
}

type Testcase struct {
	Name              string
	Description       string
	Optional          bool
	Image             string
	Port              int
	T                 TestFunc
	SetUpContainer    func(image string, runner *Runner) (*Container, error)
	TearDownContainer func(container *Container, runner *Runner)
}

type TestFunc func(port int)

type Config struct {
	Images        map[string]string
	FocusedTests  []string
	FocusedSuites []string
	Listener      ExecutionListener
	NoPull        bool
}

type Runner struct {
	config       *Config
	suites       []*Suite
	dockerClient *client.Client
}

type Container struct {
	id       string
	hostPort int
}

type runnerError error

type ExecutionListener interface {
	AboutToStart(suites []*Suite, tests map[*Suite][]*Testcase)
	SuiteStart(suite *Suite)

	TechnicalError(testcase *Testcase, result interface{})
	OptionalFailure(testcase *Testcase, result interface{})
	HardFailure(testcase *Testcase, result interface{})
	Pass(testcase *Testcase)
}

func NewRunner(config *Config, suites []*Suite) Runner {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	return Runner{config: config, suites: suites, dockerClient: cli}
}

func (r *Runner) Run() {
	requiredImages := make(map[string]struct{}, 0)

	suites := make([]*Suite, 0)
	tests := make(map[*Suite][]*Testcase)
	names := make(map[string]interface{})

	for _, s := range r.suites {
		checkSuite(s, names)
		if r.suiteShouldRun(s) {
			suites = append(suites, s)
		} else {
			continue
		}
		for _, c := range s.Cases {
			checkTest(c, names)
			if r.testShouldRun(c) {
				tests[s] = append(tests[s], c)
			} else {
				continue
			}
			requiredImages[c.Image] = struct{}{}
			if c.Port == 0 {
				c.Port = s.Port
			}
		}
	}

	for i, _ := range requiredImages {
		if r.config.Images[i] == "" {
			panic(fmt.Sprintf("Missing image for %q in config", i))
		}
	}

	if !r.config.NoPull {
		for i, _ := range requiredImages {
			if _, err := r.dockerClient.ImagePull(context.Background(), r.config.Images[i], types.ImagePullOptions{}); err != nil {
				panic(fmt.Sprintf("Error pulling image %q -> %v: %v", i, r.config.Images[i], err))
			}
		}
	}

	r.config.Listener.AboutToStart(suites, tests)

	for _, s := range suites {
		r.config.Listener.SuiteStart(s)
		for _, c := range tests[s] {
			c.Run(r)
		}
	}

}

func checkSuite(s *Suite, names map[string]interface{}) {
	if s.Name == "" {
		panic(fmt.Sprintf("Suite %v is missing a Name", s))
	} else if s.Description == "" {
		panic(fmt.Sprintf("Suite %v is missing a Description", s))
	}
	if old, ok := names[s.Name]; ok {
		panic(fmt.Sprintf("Suite %q is named like another item %v", s.Name, old))
	} else {
		names[s.Name] = s
	}
}

func checkTest(t *Testcase, names map[string]interface{}) {
	if t.Name == "" {
		panic(fmt.Sprintf("TestCase %v is missing a Name", t))
	} else if t.Description == "" {
		panic(fmt.Sprintf("TestCase %v is missing a Description", t))
	}
	if old, ok := names[t.Name]; ok {
		panic(fmt.Sprintf("Test %q is named like another item %v", t.Name, old))
	} else {
		names[t.Name] = t
	}
}

// A suite should run if a) there are no focused suites or b) the suite is focused or c) there are focused tests and they belong to that suite
func (r *Runner) suiteShouldRun(s *Suite) bool {
	if len(r.config.FocusedSuites) == 0 {
		return true
	}
	for _, fs := range r.config.FocusedSuites {
		if fs == s.Name {
			return true
		}
	}
	for _, t := range s.Cases {
		if r.isTestFocused(t) {
			return true
		}
	}
	return false
}

// A test should run (assuming its suite has already been greenlit) if a) there are no focused tests or b) it is explicitly focused
func (r *Runner) testShouldRun(t *Testcase) bool {
	if len(r.config.FocusedTests) == 0 {
		return true
	}
	return r.isTestFocused(t)
}

func (r *Runner) isTestFocused(t *Testcase) bool {
	for _, ft := range r.config.FocusedTests {
		if ft == t.Name {
			return true
		}
	}
	return false
}

func (t *Testcase) Run(runner *Runner) {
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(runnerError); ok {
				runner.config.Listener.TechnicalError(t, re)
			} else if t.Optional {
				runner.config.Listener.OptionalFailure(t, r)
			} else {
				runner.config.Listener.HardFailure(t, r)
			}
		} else {
			runner.config.Listener.Pass(t)
		}
	}()
	image := runner.config.Images[t.Image]
	var container *Container
	var err error
	if t.SetUpContainer != nil {
		container, err = t.SetUpContainer(image, runner)
	} else {
		container, err = runner.defaultSetUpContainer(image, t)
	}
	if err != nil {
		panic(runnerError(err))
	}
	defer func() {
		if container != nil {
			if t.TearDownContainer != nil {
				t.TearDownContainer(container, runner)
			} else {
				runner.defaultTearDownContainer(container)
			}
		}
	}()
	t.T(container.hostPort)
}

func (r *Runner) defaultSetUpContainer(image string, t *Testcase) (*Container, error) {
	hostPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	hostBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}
	containerPort := nat.Port(fmt.Sprintf("%d", t.Port))
	if err != nil {
		return nil, err
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := r.dockerClient.ContainerCreate(context.Background(), &container.Config{Image: image, ExposedPorts: nat.PortSet{containerPort: struct{}{}}},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, "")
	if err != nil {
		return nil, err
	}

	err = r.dockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	for i := 0; i < 10; i = i + 1 {
		req, _ := http.NewRequest("HEAD", fmt.Sprintf("http://localhost:%d/", hostPort), nil)
		_, err := http.DefaultClient.Do(req)
		if err == nil {
			break
		}
		//fmt.Printf("%2d Attempting to connect to localhost:%d: %s\n", i, hostPort, err)
		time.Sleep(20 * time.Millisecond * time.Duration(math.Pow(2, float64(i))))
	}
	if err != nil {
		return nil, err
	}

	return &Container{id: cont.ID, hostPort: hostPort}, nil
}

func (r *Runner) defaultTearDownContainer(container *Container) {
	err := r.dockerClient.ContainerKill(context.Background(), container.id, "KILL")
	if err != nil {
		panic(err)
	}
	err = r.dockerClient.ContainerRemove(context.Background(), container.id, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
