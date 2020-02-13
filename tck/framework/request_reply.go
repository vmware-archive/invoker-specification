package framework

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

var Request_reply = Suite{
	Name:        "rr",
	Description: "Request / Reply Interaction",
	Port:        8080,
	Cases: []*Testcase{
		/*
			{
				Name:           "rr-0000",
				Description:    "MUST start an http/2 server listening on $PORT",
				Image:          "upper",
				SetUpContainer: setUpContainerUsingPortEnvVar,
				TearDownContainer: func(container *Container, runner *Runner) {

				},
				T: func(port int) {
					req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
					if err != nil {
						panic(err)
					}
					req.Header.Set("Content-Type", "text/plain")
					response, err := http.DefaultClient.Do(req)
					if err != nil {
						panic(err)
					}
					if response.StatusCode != http.StatusOK {
						panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
					}
				},
			},
		*/
		{
			Name:        "rr-0001",
			Description: "MUST NOT reply on paths other than / or methods other than POST",
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/bogus", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if "HELLO" == string(result) {
					panic("The function function should only be exposed on /")
				}

				req, err = http.NewRequest("PUT", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "text/plain")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if "HELLO" == string(result) {
					panic("The function should only be exposed on /")
				}
			},
		},
		{
			Name:        "rr-0002",
			Description: "MUST honor the Accept header",
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if "HELLO" != string(result) {
					panic("Expected result as text/plain HELLO, got " + string(result))
				} else {
					hs := response.Header[http.CanonicalHeaderKey("Content-Type")]
					if len(hs) != 1 {
						panic("No Content-Type set on response")
					}
					mediaType, _, err := mime.ParseMediaType(hs[0])
					if err != nil {
						panic(fmt.Sprintf("Error parsing content-type: %v", err))
					} else if mediaType != "text/plain" {
						panic(fmt.Sprintf("Expected response Content-Type to be set to text/plain, got %v", mediaType))
					}
				}

				req, err = http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "application/json")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if `"HELLO"` != string(result) {
					panic(`Expected result as application/json "HELLO" (with quotes), got ` + string(result))
				} else {
					hs := response.Header[http.CanonicalHeaderKey("Content-Type")]
					if len(hs) != 1 {
						panic("No Content-Type set on response")
					}
					mediaType, _, err := mime.ParseMediaType(hs[0])
					if err != nil {
						panic(fmt.Sprintf("Error parsing content-type: %v", err))
					} else if mediaType != "application/json" {
						panic(fmt.Sprintf("Expected response Content-Type to be set to application/json, got %v", mediaType))
					}
				}
			},
		},
		{
			Name:        "rr-0003",
			Description: "SHOULD reply with 415 on unrecognized Content-Type",
			Optional:    true,
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "bogus/content-type")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusUnsupportedMediaType {
					panic(fmt.Sprintf("Expected 415 http code, got %d", response.StatusCode))
				}
			},
		},
		{
			Name:        "rr-0004",
			Description: "MUST reply with 5xx on unmarshalling error",
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`"hello`)) // malformed json
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode < 500 {
					panic(fmt.Sprintf("Expected 5xx http code, got %d", response.StatusCode))
				}
			},
		},
		{
			Name:        "rr-0005",
			Description: "SHOULD reply with 406 on inability to marshall back",
			Optional:    true,
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`hello`)) // malformed json
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "not/gonna-happen")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusNotAcceptable {
					panic(fmt.Sprintf("Expected 406 http code, got %d", response.StatusCode))
				}
			},
		},
		{
			Name:        "rr-0006",
			Description: "MUST survive invocation errors",
			Optional:    true,
			Image:       "divider",
			T: func(port int) {
				// Make sure function works first
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`2`))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if `50` != string(result) {
					panic(`Expected result as application/json 50, got ` + string(result))
				}
				// Trigger an invocation error
				req, err = http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`0`))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusInternalServerError {
					panic(fmt.Sprintf(`Expected http status 500, got %d`, response.StatusCode))
				}
				// Verify error above did not crash the whole process
				req, err = http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`4`))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if `25` != string(result) {
					panic(`Expected result as application/json 25, got ` + string(result))
				}
			},
		},
		{
			Name:        "rr-0007",
			Description: "MAY support functions that maintain state",
			Optional:    true,
			Image:       "counter",
			T: func(port int) {
				wg := sync.WaitGroup{}
				wg.Add(99)
				errors := make(chan interface{}, 100)
				f := func(c int, errs chan interface{}) int {
					req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(fmt.Sprintf("%d", c)))
					if err != nil {
						panic(err)
					}
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Accept", "application/json")
					response, err := http.DefaultClient.Do(req)
					if err != nil {
						errs <- err
						return -1
					}
					if response.StatusCode != http.StatusOK {
						errs <- fmt.Sprintf("Expected 200 http code, got %d", response.StatusCode)
						return -1
					}
					if result, err := ioutil.ReadAll(response.Body); err != nil {
						errs <- err
					} else if s, err := strconv.Atoi(string(result)); err != nil {
						errs <- err
					} else {
						return s
					}
					return -1
				}

				for i := 1; i <= 99; i++ {
					go func(s int, e chan interface{}) {
						defer wg.Done()
						f(s, e)
					}(i, errors)

				}
				wg.Wait()
				select {
				case e := <-errors:
					panic(e)
				default:
				}

				result := f(100, errors)
				select {
				case e := <-errors:
					panic(e)
				default:
				}
				if result != 100*101/2 {
					panic(fmt.Sprintf("Expected invocations of counter with values 1..100 to sum up to %v, got %v", 100*101/2, result))
				}
			},
		},
		{
			Name:        "rr-0008",
			Description: "MUST assume application/octet-stream when no Content-Type",
			Image:       "md5",
			T: func(port int) {
				// Make sure function works first
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Accept", "application/json")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if `"5d41402abc4b2a76b9719d911017c592"` != string(result) {
					panic(`Expected result as application/json "5d41402abc4b2a76b9719d911017c592" (with quotes), got ` + string(result))
				}
			},
		},
		{
			Name:        "rr-0009",
			Description: "MUST assume Accept: */* when no Accept set",
			Image:       "upper",
			T: func(port int) {
				// Make sure function works first
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if hs := response.Header[http.CanonicalHeaderKey("Content-Type")]; len(hs) == 1 {
					mediaType, _, err := mime.ParseMediaType(hs[0])
					if err != nil {
						panic(fmt.Sprintf("Error parsing Content-Type header %v: %v", hs[0], err))
					}
					if mediaType == "text/plain" && string(result) != "HELLO" {
						panic("Advertised response as text/plain but did not get expected HELLO")
					} else if mediaType == "application/json" && string(result) != `"HELLO"` {
						panic(`Advertised response as application/json but did not get expected "HELLO"`)
					} else if mediaType == "application/octet-stream" && string(result) != "HELLO" {
						panic(`Advertised response as application/octet-stream but did not get expected HELLO`)
					} else {
						// Can't be sure this is wrong. Simply return quietly
					}
				} else {
					panic("Response Content-Type should have been set to a single value")
				}
			},
		},
	},
}

func setUpContainerUsingPortEnvVar(image string, r *Runner) (*Container, error) {
	_, err := r.dockerClient.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	hostPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	hostBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}
	containerPort := nat.Port("4321")
	if err != nil {
		return nil, err
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := r.dockerClient.ContainerCreate(context.Background(),
		&container.Config{Image: image, ExposedPorts: nat.PortSet{containerPort: struct{}{}}, Env: []string{"PORT=4321"}},
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
		_, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", hostPort))
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond * time.Duration(math.Pow(2, float64(i))))
	}
	if err != nil {
		return nil, err
	}
	// TODO: need to sleep some more. Find a more reliable way to diagnose a container as ready
	time.Sleep(1000 * time.Millisecond)

	return &Container{id: cont.ID, hostPort: hostPort}, nil
}
