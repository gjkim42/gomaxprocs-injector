package admission

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wI2L/jsondiff"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAdmit(t *testing.T) {
	testCases := []struct {
		desc   string
		review v1.AdmissionReview

		allowed                          bool
		expectedInitContainersGOMAXPROCS []int
		expectedContainersGOMAXPROCS     []int
	}{
		{
			desc: "should not accept a resource other than pods",
			review: v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Resource: metav1.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "configmaps",
					},
				},
			},
			allowed: false,
		},
		{
			desc: "test pod",
			review: v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Resource: metav1.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					},
					Object: newTestPodObject(t,
						"test-pod",
						corev1.Container{
							Name: "container-without-cpu-limit",
						},
						containerWithCPULimit("100m"),
						containerWithCPULimit("1"),
						containerWithCPULimit("1100m"),
						containerWithCPULimit("1900m"),
						containerWithCPULimit("2"),
						containerWithCPULimit("2500m"),
						corev1.Container{
							Name: "container-with-GOMAXPROCS",
							Env: []corev1.EnvVar{
								{
									Name:  "GOMAXPROCS",
									Value: "3",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("8"),
								},
							},
						},
					),
				},
			},
			allowed: true,
			expectedContainersGOMAXPROCS: []int{
				0, // container-without-cpu-limit
				1,
				1,
				2,
				2,
				2,
				3,
				3, // container-with-GOMAXPROCS
			},
		},
		{
			desc: "test pod with init containers",
			review: v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Resource: metav1.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					},
					Object: newPodObjectFromPod(t, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod-with-init-containers",
							Namespace: "default",
						},
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								containerWithCPULimit("100m"),
							},
							Containers: []corev1.Container{
								containerWithCPULimit("3"),
							},
						},
					}),
				},
			},
			allowed: true,
			expectedInitContainersGOMAXPROCS: []int{
				1,
			},
			expectedContainersGOMAXPROCS: []int{
				3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run("v1 "+tc.desc, func(t *testing.T) {
			c := NewController()
			res := c.admit(tc.review)
			if res.Allowed != tc.allowed {
				t.Errorf("expected %v, got %v", tc.allowed, res.Allowed)
			}
			if tc.allowed {
				checkGOMAXPROCS(t,
					tc.review.Request.Object.Raw,
					res.Patch,
					tc.expectedInitContainersGOMAXPROCS,
					tc.expectedContainersGOMAXPROCS)
			}
		})
		t.Run("v1beta1 "+tc.desc, func(t *testing.T) {
			c := NewController()
			res := c.admitV1beta1(v1beta1.AdmissionReview{
				Request: convertAdmissionRequestToV1beta1(tc.review.Request),
			})
			if res.Allowed != tc.allowed {
				t.Errorf("expected %v, got %v", tc.allowed, res.Allowed)
			}
			// TODO: check GOMAXPROCS
		})
	}
}

func newTestPodObject(t *testing.T, name string, containers ...corev1.Container) runtime.RawExtension {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}

	return newPodObjectFromPod(t, pod)
}

func newPodObjectFromPod(t *testing.T, pod *corev1.Pod) runtime.RawExtension {
	rawBytes, err := json.Marshal(pod)
	if err != nil {
		t.Fatal(err)
	}

	return runtime.RawExtension{
		Raw: rawBytes,
	}
}

func containerWithCPULimit(cpu string) corev1.Container {
	return corev1.Container{
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse(cpu),
			},
		},
	}
}

func checkGOMAXPROCS(t *testing.T, rawObject, patch []byte, expectedInitContainersGOMAXPROCS, expectedContainersGOMAXPROCS []int) {
	var pod corev1.Pod
	if err := json.Unmarshal(rawObject, &pod); err != nil {
		t.Fatal(err)
	}

	expectedPod := pod.DeepCopy()
	for i := range expectedPod.Spec.InitContainers {
		expectedPod.Spec.InitContainers[i].Env = applyGOMAXPROCSToEnv(expectedPod.Spec.InitContainers[i].Env, expectedInitContainersGOMAXPROCS[i])
	}

	for i := range expectedPod.Spec.Containers {
		expectedPod.Spec.Containers[i].Env = applyGOMAXPROCSToEnv(expectedPod.Spec.Containers[i].Env, expectedContainersGOMAXPROCS[i])
	}

	expectedPatch, err := jsondiff.Compare(&pod, expectedPod)
	if err != nil {
		t.Fatal(err)
	}

	expectedPatchBytes, err := json.Marshal(expectedPatch)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(expectedPatchBytes))
	t.Log(string(patch))

	if diff := cmp.Diff(expectedPatchBytes, patch); diff != "" {
		t.Errorf("unexpected patch (-want +got):\n%s", diff)
	}

	return
}

func applyGOMAXPROCSToEnv(env []corev1.EnvVar, gomaxprocs int) []corev1.EnvVar {
	if gomaxprocs == 0 {
		return env
	}

	for i, envVar := range env {
		if envVar.Name == "GOMAXPROCS" {
			env[i].Value = strconv.Itoa(gomaxprocs)
			return env
		}
	}

	return append(env, corev1.EnvVar{
		Name:  "GOMAXPROCS",
		Value: strconv.Itoa(gomaxprocs),
	})
}
