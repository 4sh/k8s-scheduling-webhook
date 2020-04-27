// Package mutate is responsible for adding toleration to the requested pod creation.
package mutate

import (
	"encoding/json"
	"fmt"
	"github.com/mattbaird/jsonpatch"
	"log"
	"os"
	"strings"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Mutate(body []byte, injectionId string, verbose bool) ([]byte, error) {
	if verbose {
		log.Printf("recv: %s\n", string(body))
	}

	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar != nil {
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}

		resp.Allowed = true
		resp.UID = ar.UID
		pT := v1beta1.PatchTypeJSONPatch
		resp.PatchType = &pT

		resp.AuditAnnotations = map[string]string{
			"toleration-injection-for": injectionId,
		}

		tolerations, err := GetTolerationsConfig(injectionId)
		if err != nil {
			return nil, fmt.Errorf("unable get toleration config for %s - %v", injectionId, err)
		}

		var patch []jsonpatch.JsonPatchOperation
		patch = append(patch, jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      "/spec/tolerations",
			Value:     tolerations,
		})

		resp.Patch, err = json.Marshal(patch)

		resp.Result = &metav1.Status{Status: "Success"}

		admReview.Response = &resp

		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err // untested section
		}
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody)) // untested section
	}

	return responseBody, nil
}

func GetTolerationsConfig(injectionId string) ([]corev1.Toleration, error) {
	var tolerationKey = os.Getenv(fmt.Sprintf("TOLERATION_KEY_%s", strings.ToUpper(injectionId)))
	var tolerationValue = os.Getenv(fmt.Sprintf("TOLERATION_VALUE_%s", strings.ToUpper(injectionId)))
	var tolerationEffect = corev1.TaintEffect(os.Getenv(fmt.Sprintf("TOLERATION_EFFECT_%s", strings.ToUpper(injectionId))))

	var tolerations = []corev1.Toleration{
		{Key: tolerationKey, Value: tolerationValue, Effect: tolerationEffect},
	}

	return tolerations, nil
}
