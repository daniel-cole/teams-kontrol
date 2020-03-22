package command

import (
	"fmt"
	"github.com/daniel-cole/teams-kontrol/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	teamsKontrolPermissionFile := "testdata/permissions.yml"
	err := util.AttemptSetEnv(KontrolPermissionFileEnvKey, teamsKontrolPermissionFile)
	if err != nil {
		log.Fatalf("Failed to set %s", KontrolPermissionFileEnvKey)
	}
	Init()
	os.Exit(m.Run())
}

// test valid Command with no identifier
func TestParseAndValidateGetPodCommandValid(t *testing.T) {
	// valid Command: get pods default
	validCommand := "get pods default"
	command, err := parseAndValidateCommandFromString(validCommand)
	if err != nil {
		t.Fatalf("expected Command to be valid: %s", validCommand)
	}
	expectedVerb := "get"
	if command.Verb != expectedVerb {
		t.Errorf("expected verb to be: %s", expectedVerb)
	}

	expectedResource := "pods"
	if expectedResource != command.Resource {
		t.Fatalf("expected resource to be %s", expectedResource)
	}

	expectedNamespace := "default"
	if expectedNamespace != command.Namespace {
		t.Errorf("expected namespace to be %s", expectedNamespace)
	}

}

// test with identifier
func TestParseAndValidateGetPodCommandValidIdent(t *testing.T) {
	identifier := "redis-asdqwe-23dd2"
	validCommand := fmt.Sprintf("describe pods redis %s", identifier)
	command, err := parseAndValidateCommandFromString(validCommand)
	if err != nil {
		t.Fatalf("expected Command to be valid: %s", validCommand)
	}

	expectedVerb := "describe"
	if command.Verb != expectedVerb {
		t.Errorf("expected verb to be: %s", expectedVerb)
	}

	expectedResource := "pods"
	if expectedResource != command.Resource {
		t.Fatalf("expected resource to be %s", expectedResource)
	}

	expectedNamespace := "redis"
	if expectedNamespace != command.Namespace {
		t.Errorf("expected namespace to be %s", expectedNamespace)
	}

	if identifier != command.Identifier {
		t.Fatalf("expected identifier to be %s, instead got %s", identifier, command.Identifier)
	}
}

func TestParseAndValidatePodCommandInvalid(t *testing.T) {
	// valid Command: get pods default
	invalidCommand := "get pox default"
	_, err := parseAndValidateCommandFromString(invalidCommand)
	if err == nil {
		t.Fatalf("expected Command to be invalid: %s", invalidCommand)
	}
}

func TestExecuteGetPodsCommand(t *testing.T) {

	client := fake.NewSimpleClientset()

	namespace := "nginx"
	image := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	podNames := []string{"nginx-ingress-controller-bsdad", "nginx-ingress-controller-a12fb"}
	for _, name := range podNames {
		err := createSimplePod(client, name, namespace, image)
		if err != nil {
			t.Fatalf("failed to create pod: %v", err)
		}
	}

	command, err := parseAndValidateCommandFromString("get pods nginx")
	if err != nil {
		t.Fatalf("failed to parse and validate command")
	}
	result, err := ExecuteCommand(client, command)
	if err != nil {
		t.Fatalf("failed to execute Command: %s. %v", command, err)
	}

	switch resultType := result.(type) {
	case *v1.PodList:
	default:
		t.Fatalf("got unexpected result type: %v", resultType)
	}

	castResult := result.(*v1.PodList)

	totalPodsFound := len(castResult.Items)
	expectedPodsFound := 2
	if totalPodsFound != expectedPodsFound {
		t.Fatalf("expected results to contain %d items, instead got %d", expectedPodsFound, totalPodsFound)
	}

	// create a list of the pod names that were found
	var foundPods []string
	for _, item := range castResult.Items {
		foundPods = append(foundPods, item.Name)
	}

	// check that both pods are returned correctly
	for _, name := range podNames {
		if !util.StringInSliceIgnoreCase(name, foundPods) {
			t.Fatalf("expected to find %s in results", name)
		}
	}

}

func TestExecuteGetPodCommand(t *testing.T) {

	client := fake.NewSimpleClientset()

	namespace := "nginx"
	image := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	podName := "nginx-ingress-controller-a12fb"
	err := createSimplePod(client, podName, namespace, image)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	command, err := parseAndValidateCommandFromString("get pod nginx nginx-ingress-controller-a12fb")
	if err != nil {
		t.Fatalf("failed to parse and validate command")
	}
	result, err := ExecuteCommand(client, command)
	if err != nil {
		t.Fatalf("failed to execute Command: %s. %v", command, err)
	}

	switch resultType := result.(type) {
	case *v1.Pod:
	default:
		t.Fatalf("got unexpected result type: %v", resultType)
	}

}

// UUID is randomly generated it makes it difficult to test the output
// maybe work on testing this more exhaustively later
func TestRenderPodTeamsCard(t *testing.T) {

	name := "nginx"
	namespace := "nginx"
	image := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Image: image,
			}},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	_, err := renderTeamsPodCard([]v1.Pod{pod})
	if err != nil {
		t.Fatalf("failed to render teams card for pod: %v", err)
	}

	//fmt.Println(strconv.Unquote(string(out))) //useful for dumping formatted json

}

func TestRenderPodListTeamsCard(t *testing.T) {
	name1 := "nginx-1"
	namespace1 := "nginx"
	image1 := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	pod1 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name1,
			Namespace: namespace1,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Image: image1,
			}},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	name2 := "nginx-2"
	namespace2 := "nginx"
	image2 := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	pod2 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name2,
			Namespace: namespace2,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Image: image2,
			}},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	podList := []v1.Pod{pod1, pod2}

	_, err := renderTeamsPodCard(podList)
	if err != nil {
		t.Fatalf("failed to render teams card for pod: %v", err)
	}
}

func TestDeletePod(t *testing.T) {

	client := fake.NewSimpleClientset()

	namespace := "nginx"
	image := "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.25.0"

	podName := "nginx-ingress-controller-a12fb"
	err := createSimplePod(client, podName, namespace, image)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	command, err := parseAndValidateCommandFromString("delete pod nginx " + podName)
	if err != nil {
		t.Fatalf("failed to parse and validate command: %v", err)
	}
	_, err = ExecuteCommand(client, command)
	if err != nil {
		t.Fatalf("failed to execute Command: %s. %v", command, err)
	}

}

func createSimplePod(client kubernetes.Interface, name string, namespace string, image string) error {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Image: image,
			}},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}
	_, err := client.CoreV1().Pods(namespace).Create(pod)
	if err != nil {
		return err
	}
	return nil
}
