package platform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"k8s.io/client-go/deprecated/scheme"
	testcore "k8s.io/client-go/testing"

	"k8s.io/kubernetes/pkg/util/yaml"

	"github.com/alcounit/selenosis/selenium"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
	yayaml "gopkg.in/yaml.v3"
	"gotest.tools/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestErrorsOnServiceCreate(t *testing.T) {
	tests := map[string]struct {
		ns        string
		podName   string
		layout    ServiceSpec
		eventType watch.EventType
		podPhase  apiv1.PodPhase
		err       error
	}{
		"Verify platform error on pod startup phase PodSucceeded": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodSucceeded,
			err:       errors.New("pod is not ready after creation: pod exited early with status Succeeded"),
		},
		"Verify platform error on pod startup phase PodFailed": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodFailed,
			err:       errors.New("pod is not ready after creation: pod exited early with status Failed"),
		},
		"Verify platform error on pod startup phase PodUnknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: couldn't obtain pod state"),
		},
		"Verify platform error on pod startup phase Unknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			err:       errors.New("pod is not ready after creation: pod has unknown status"),
		},
		"Verify platform error on pod startup event Error": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Error,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: received error while watching pod: /, Kind="),
		},
		"Verify platform error on pod startup event Deleted": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Deleted,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: pod was deleted before becoming available"),
		},
		"Verify platform error on pod startup event Unknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Bookmark,
			podPhase:  apiv1.PodSucceeded,
			err:       errors.New("pod is not ready after creation: received unknown event type BOOKMARK while watching pod"),
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()
		watcher := watch.NewFakeWithChanSize(1, false)
		mock.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
		watcher.Action(test.eventType, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: test.podName,
			},
			Status: apiv1.PodStatus{
				Phase: test.podPhase,
			},
		})

		client := &Client{
			ns:        test.ns,
			clientset: mock,
			service: &service{
				ns:        test.ns,
				clientset: mock,
			},
		}

		_, err := client.Service().Create(test.layout)

		assert.Equal(t, test.err.Error(), err.Error())
	}
}

func TestPodDelete(t *testing.T) {
	tests := map[string]struct {
		ns           string
		createPod    string
		deletePod    string
		browserImage string
		proxyImage   string
		err          error
	}{
		"Verify platform deletes running pod": {
			ns:           "selenosis",
			createPod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			deletePod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
		},
		"Verify platform delete return error": {
			ns:           "selenosis",
			createPod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			deletePod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912",
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
			err:          errors.New(`pods "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912" not found`),
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()

		client := &Client{
			ns:        test.ns,
			clientset: mock,
			service: &service{
				ns:        test.ns,
				clientset: mock,
			},
		}

		ctx := context.Background()
		_, err := mock.CoreV1().Pods(test.ns).Create(ctx, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: test.createPod,
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "browser",
						Image: test.browserImage,
					},
					{
						Name:  "seleniferous",
						Image: test.proxyImage,
					},
				},
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create fake pod: %v", err)
		}

		err = client.Service().Delete(test.deletePod)

		if err != nil {
			assert.Equal(t, test.err.Error(), err.Error())
		} else {
			assert.Equal(t, test.err, err)
		}
	}

}

func TestListPods(t *testing.T) {
	tests := map[string]struct {
		ns       string
		svc      string
		podNames []string
		podData  []struct {
			podName   string
			podPhase  apiv1.PodPhase
			podStatus ServiceStatus
		}
		podPhase     []apiv1.PodPhase
		podStatus    []ServiceStatus
		labels       map[string]string
		browserImage string
		proxyImage   string
		err          error
	}{
		"Verify platform returns running pods that match selector": {
			ns:           "selenosis",
			svc:          "selenosis",
			podNames:     []string{"chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911", "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912", "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144913"},
			podPhase:     []apiv1.PodPhase{apiv1.PodRunning, apiv1.PodPending, apiv1.PodFailed},
			podStatus:    []ServiceStatus{Running, Pending, Unknown},
			labels:       map[string]string{"selenosis.app.type": "browser"},
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()
		client := &Client{
			ns:        test.ns,
			svc:       test.svc,
			svcPort:   intstr.FromString("4445"),
			clientset: mock,
		}

		for i, name := range test.podNames {
			ctx := context.Background()
			_, err := mock.CoreV1().Pods(test.ns).Create(ctx, &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: test.labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "browser",
							Image: test.browserImage,
						},
						{
							Name:  "seleniferous",
							Image: test.proxyImage,
						},
					},
				},
				Status: apiv1.PodStatus{
					Phase: test.podPhase[i],
				},
			}, metav1.CreateOptions{})

			if err != nil {
				t.Logf("failed to create fake pod: %v", err)
			}
		}

		state, err := client.State()
		if err != nil {
			t.Fatalf("Failed to list pods %v", err)
		}

		for i, pod := range state.Services {
			assert.Equal(t, pod.SessionID, test.podNames[i])

			u := &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(fmt.Sprintf("%s.%s", test.podNames[i], test.svc), "4445"),
			}

			assert.Equal(t, pod.URL.String(), u.String())

			assert.Equal(t, pod.Status, test.podStatus[i])
		}
	}
}

func TestGetPodManifest(t *testing.T) {
	cfgFile, _ := filepath.Abs("./../config/browsers.yaml")
	//t.Logf("Try to read config  %s", cfgFile)

	layouts, err := readConfig(cfgFile, t)
	if err != nil {
		t.Logf("Config not read: %v", err)
		return
	}
	t.Logf("Layouts %v", layouts)

	type capabilities struct {
		DesiredCapabilities selenium.Capabilities `json:"desiredCapabilities"`
		Capabilities        struct {
			AlwaysMatch selenium.Capabilities    `json:"alwaysMatch"`
			FirstMatch  []*selenium.Capabilities `json:"firstMatch"`
		} `json:"capabilities"`
	}

	request := capabilities{}

	firstMatchCaps := request.Capabilities.FirstMatch
	if len(firstMatchCaps) == 0 {
		firstMatchCaps = append(firstMatchCaps, &selenium.Capabilities{})
	}

	var browser BrowserSpec
	var caps selenium.Capabilities

	for _, fmc := range firstMatchCaps {
		caps = request.DesiredCapabilities
		mergo.Merge(&caps, *fmc)
		caps.ValidateCapabilities()

		browser, err = FindContainer(layouts, caps.GetBrowserName(), caps.BrowserVersion)
		if err != nil {
			t.Logf("unknown browser %s", err)
		}
	}

	if err != nil {
		t.Logf("requested browser not found: %v", err)
		return
	}

	image := parseImage(browser.Image)
	var svcPort intstr.IntOrString
	svcPort = intstr.FromString("4445")
	pod := getPodManifest(ServiceSpec{
		SessionID:             fmt.Sprintf("%s-%s", image, uuid.New()),
		RequestedCapabilities: caps,
		Template:              browser,
	}, "svc", "proxyImage", svcPort, "idleTimeout", "ns", "imagePullSecretName")

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	body := runtime.EncodeOrDie(codec, pod)
	t.Logf("pod dump: \n%s", body)

	/*******************             handlers        ****************************/
}

func parseImage(image string) (container string) {
	if len(image) > 0 {
		pref, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			return "browser"
		}
		fragments := strings.Split(image, "/")
		image = fragments[len(fragments)-1]
		return pref.ReplaceAllString(image, "-")
	}
	return "browser"
}

type Layout struct {
	DefaultSpec     Spec                    `yaml:"spec" json:"spec"`
	Meta            Meta                    `yaml:"meta" json:"meta"`
	Path            string                  `yaml:"path" json:"path"`
	DefaultVersion  string                  `yaml:"defaultVersion" json:"defaultVersion"`
	Versions        map[string]*BrowserSpec `yaml:"versions" json:"versions"`
	Volumes         []apiv1.Volume          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Capabilities    []apiv1.Capability      `yaml:"kernelCaps,omitempty" json:"kernelCaps,omitempty"`
	RunAs           RunAsOptions            `yaml:"runAs,omitempty" json:"runAs,omitempty"`
	SecurityContext apiv1.SecurityContext   `yaml:"securityContext,omitempty" json:"securityContext,omitempty"`
}

func readConfig(configFile string, t *testing.T) (map[string]*Layout, error) {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read error: %v", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(content), 1000)

	layouts := make(map[string]*Layout)
	if err := decoder.Decode(&layouts); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}

	if len(layouts) == 0 {
		return nil, fmt.Errorf("empty config: %v", err)
	}

	for _, layout := range layouts {
		spec := layout.DefaultSpec
		for _, container := range layout.Versions {
			if container.Path == "" {
				container.Path = layout.Path
			}
			container.Meta.Annotations = merge(container.Meta.Annotations, layout.Meta.Annotations)
			container.Meta.Labels = merge(container.Meta.Labels, layout.Meta.Labels)
			container.Volumes = layout.Volumes
			container.Capabilities = append(container.Capabilities, layout.Capabilities...)

			if err := mergo.Merge(&container.SecurityContext, layout.SecurityContext); err != nil {
				return nil, fmt.Errorf("merge error %v", err)
			}

			if err := mergo.Merge(&container.Spec, spec); err != nil {
				return nil, fmt.Errorf("merge error %v", err)
			}

			if err := mergo.Merge(&container.RunAs, layout.RunAs); err != nil {
				return nil, fmt.Errorf("merge error %v", err)
			}
		}
	}

	d, err := yayaml.Marshal(&layouts)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	return layouts, nil
}

func FindContainer(layouts map[string]*Layout, name string, version string) (BrowserSpec, error) {
	c, ok := layouts[name]
	if !ok {
		return BrowserSpec{}, fmt.Errorf("unknown browser name %s", name)
	}

	v, ok := c.Versions[version]

	if !ok {
		if c.DefaultVersion != "" {
			v, ok = c.Versions[c.DefaultVersion]
			if !ok {
				return BrowserSpec{}, fmt.Errorf("unknown browser version %s", version)
			}
			v.BrowserName = name
			v.BrowserVersion = c.DefaultVersion
			return *v, nil
		}
		return BrowserSpec{}, fmt.Errorf("unknown browser version %s", version)
	}
	v.BrowserName = name
	v.BrowserVersion = version
	return *v, nil
}

func merge(from, to map[string]string) map[string]string {
	for k, v := range from {
		to[k] = v
	}
	return to
}
