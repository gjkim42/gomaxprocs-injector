package admission

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/wI2L/jsondiff"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

var patchTypeJSONPatch = v1.PatchTypeJSONPatch

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.ErrorS(nil, "Got wrong contentType", "got", contentType, "expect", "application/json")
		return
	}

	klog.V(2).Info("Handling", "request", string(body))

	deserializer := codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to deserialize request object: %v", err)
		klog.ErrorS(err, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			klog.ErrorS(nil, "Wrong AdmissionReview type", "expect", "v1beta1.AdmissionReview", "got", fmt.Sprintf("%T", obj))
			return
		}
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = c.admitV1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			klog.ErrorS(nil, "Wrong AdmissionReview type", "expect", "v1.AdmissionReview", "got", fmt.Sprintf("%T", obj))
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = c.admit(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.ErrorS(nil, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info("Sending", "response", responseObj)
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		msg := fmt.Sprintf("Failed to serialize response object: %v", err)
		klog.ErrorS(err, msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.ErrorS(err, "Failed to write response")
	}
}

func (c *Controller) admit(review v1.AdmissionReview) *v1.AdmissionResponse {
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if review.Request.Resource != podResource {
		err := fmt.Errorf("expected resource to be %s", podResource)
		klog.ErrorS(err, "Failed to admit")
		return toV1AdmissionResponse(err)
	}

	var pod corev1.Pod
	if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
		klog.ErrorS(err, "Failed to unmarshal pod")
		return toV1AdmissionResponse(err)
	}

	klog.InfoS("Admitting a pod", "pod", klog.KObj(&pod))

	newPod := pod.DeepCopy()
	for i := range newPod.Spec.InitContainers {
		if err := mutateContainer(&newPod.Spec.InitContainers[i]); err != nil {
			klog.ErrorS(err, "Failed to mutate container")
			return toV1AdmissionResponse(err)
		}
	}
	for i := range newPod.Spec.Containers {
		if err := mutateContainer(&newPod.Spec.Containers[i]); err != nil {
			klog.ErrorS(err, "Failed to mutate container")
			return toV1AdmissionResponse(err)
		}
	}

	patch, err := jsondiff.Compare(pod, newPod)
	if err != nil {
		klog.ErrorS(err, "Failed to create JSONPatch")
		return toV1AdmissionResponse(err)
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		klog.ErrorS(err, "Failed to marshal JSONPatch")
		return toV1AdmissionResponse(err)
	}

	return &v1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchTypeJSONPatch,
	}
}

func mutateContainer(container *corev1.Container) error {
	for _, env := range container.Env {
		if env.Name == "GOMAXPROCS" {
			klog.InfoS("Container already has GOMAXPROCS set", "container", container.Name)
			return nil
		}
	}

	if container.Resources.Limits == nil || container.Resources.Limits.Cpu().IsZero() {
		klog.InfoS("Container has no cpu resource limit", "container", container.Name)
		return nil
	}

	roundedCPULimit := container.Resources.Limits.Cpu().Value()
	if roundedCPULimit < 1 {
		roundedCPULimit = 1
	}

	klog.InfoS("Setting GOMAXPROCS", "container", container.Name, "value", roundedCPULimit)

	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "GOMAXPROCS",
		Value: strconv.FormatInt(roundedCPULimit, 10),
	})

	return nil
}

func (c *Controller) admitV1beta1(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	in := v1.AdmissionReview{Request: convertAdmissionRequestToV1(review.Request)}
	out := c.admit(in)
	return convertAdmissionResponseToV1beta1(out)
}
