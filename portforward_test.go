package portforward

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
	"os"
	"os/user"
	"testing"
)

func TestPortForwardGetKubeConfigPath(t *testing.T) {
	pf := PortForward{}

	path, err := pf.getKubeConfigPath()
	assert.Nil(t, err)

	user, _ := user.Current()
	assert.Equal(t, user.HomeDir+"/.kube/config", path)
}

func TestPortForwardGetKubeConfigPathEnvVarSet(t *testing.T) {
	os.Setenv("KUBECONFIG", "/my/kube/config")
	defer os.Setenv("KUBECONFIG", "")

	pf := PortForward{}

	path, err := pf.getKubeConfigPath()
	assert.Nil(t, err)

	assert.Equal(t, "/my/kube/config", path)
}

func newPod(name string, labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
			Name:   name,
		},
	}
}

func TestFindPodByLabels(t *testing.T) {
	pf := PortForward{
		Clientset: fakekubernetes.NewSimpleClientset(
			newPod("mypod1", map[string]string{
				"name": "other",
			}),
			newPod("mypod2", map[string]string{
				"name": "flux",
			}),
			newPod("mypod3", map[string]string{})),
		Labels: map[string]string{
			"name": "flux",
		},
	}

	pod, err := pf.findPodByLabels()
	assert.Nil(t, err)
	assert.Equal(t, "mypod2", pod)
}

func TestFindPodByLabelsNoneExist(t *testing.T) {
	pf := PortForward{
		Clientset: fakekubernetes.NewSimpleClientset(
			newPod("mypod1", map[string]string{
				"name": "other",
			})),
		Labels: map[string]string{
			"name": "flux",
		},
	}

	_, err := pf.findPodByLabels()
	assert.NotNil(t, err)
	assert.Equal(t, "Could not find pod for selector: labels map[name:flux]", err.Error())
}

func TestFindPodByLabelsMultiple(t *testing.T) {
	pf := PortForward{
		Clientset: fakekubernetes.NewSimpleClientset(
			newPod("mypod1", map[string]string{
				"name": "flux",
			}),
			newPod("mypod2", map[string]string{
				"name": "flux",
			}),
			newPod("mypod3", map[string]string{})),
		Labels: map[string]string{
			"name": "flux",
		},
	}

	_, err := pf.findPodByLabels()
	assert.NotNil(t, err)
	assert.Equal(t, "Ambiguous pod: found more than one pod for selector: labels map[name:flux]", err.Error())
}

func TestGetPodNameNameSet(t *testing.T) {
	pf := PortForward{
		Name: "hello",
	}

	pod, err := pf.getPodName()
	assert.Nil(t, err)
	assert.Equal(t, "hello", pod)
}

func TestGetPodNameNoNameSet(t *testing.T) {
	pf := PortForward{
		Clientset: fakekubernetes.NewSimpleClientset(
			newPod("mypod", map[string]string{
				"name": "flux",
			})),
		Labels: map[string]string{
			"name": "flux",
		},
	}

	pod, err := pf.getPodName()
	assert.Nil(t, err)
	assert.Equal(t, "mypod", pod)
	assert.Equal(t, pf.Name, pod)
}

func TestGetFreePort(t *testing.T) {
	pf := PortForward{}
	port, err := pf.getFreePort()
	assert.Nil(t, err)
	assert.NotZero(t, port)
}

func TestGetListenPort(t *testing.T) {
	pf := PortForward{
		ListenPort: 80,
	}

	port, err := pf.getListenPort()
	assert.Nil(t, err)
	assert.Equal(t, 80, port)
}

func TestGetListenPortRandom(t *testing.T) {
	pf := PortForward{}

	port, err := pf.getListenPort()
	assert.Nil(t, err)
	assert.NotZero(t, port)
	assert.Equal(t, pf.ListenPort, port)
}
