package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daniel-cole/teams-kontrol/config"
	"github.com/daniel-cole/teams-kontrol/k8s"
	"github.com/daniel-cole/teams-kontrol/middleware"
	"github.com/daniel-cole/teams-kontrol/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type Command struct {
	Verb       string
	Resource   string
	Namespace  string
	Identifier string
}

const KontrolPermissionFileEnvKey = "TEAMS_KONTROL_PERMISSION_FILE"
const KontrolInsecureCommandHandlerEnvKey = "TEAMS_KONTROL_INSECURE_COMMANDS"
const KontrolResponseTypeEnvKey = "TEAMS_KONTROL_RESPONSE_TYPE"
const verbPermissionIndex = 0
const resourcePermissionIndex = 1
const namespacePermissionIndex = 2
const teamsResponseType = "TEAMS"

var responseType string
var permissions config.Permissions

// Init will ensure that there's a Command file. The program will crash if it is not specified.
func Init() {
	var permissionFile string
	if permissionFile = os.Getenv(KontrolPermissionFileEnvKey); permissionFile == "" {
		logrus.Fatalf("Exiting. Please specify a permissions file with %s", KontrolPermissionFileEnvKey)
	}
	permissionsFromFile, err := ioutil.ReadFile(permissionFile)
	if err != nil {
		logrus.Fatalf("Failed to read Command file: %v", err)
	}
	err = yaml.Unmarshal(permissionsFromFile, &permissions)
	if err != nil {
		logrus.Fatalf("Failed to unmarshal json for Command file %v", err)
	}

	if responseType = os.Getenv(KontrolResponseTypeEnvKey); responseType == "" {
		responseType = teamsResponseType // default to teams
	}
}

func Handler(client kubernetes.Interface, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to read request body: %v", r.Body)
		middleware.LogWithContext(ctx).Error(errorMsg)
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	commandStr := string(body)
	command, err := parseAndValidateCommandFromString(commandStr)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to parse and validate command: '%s'", commandStr)
		middleware.LogWithContext(ctx).Error(errorMsg)
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	result, err := ExecuteCommand(client, command)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to execute command: %s, got %v", commandStr, err)
		middleware.LogWithContext(ctx).Error(errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

	response, err := PrepareResponse(result)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to prepare response: %v", err)
		middleware.LogWithContext(ctx).Error(errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)

}

// PrepareResponse takes an interface and attempts to render the appropriate response and returns it as a byte array
// i.e. if teamsResponseType == "TEAMS" and the interface is a pod then it will return a teams card with the pod details
func PrepareResponse(result interface{}) ([]byte, error) {
	switch responseType {
	case teamsResponseType:
		switch castResult := result.(type) {
		case *v1.Pod:
			return renderTeamsPodCard([]v1.Pod{*castResult})
		case *v1.PodList:
			return renderTeamsPodCard(castResult.Items)
		case nil:
			return []byte("ok"), nil // if nil then we presume that the command was executed successfully
		default:
			return nil, errors.New(fmt.Sprintf("unknown type returned from execute command: %s", reflect.TypeOf(castResult)))
		}
	default:
		return nil, errors.New("unknown response type: " + responseType)
	}
}

// Execute takes a valid command and attempts to execute it
// returns an interface containing a list of pods, a pod, or an error if it's failed.
// if nil, nil is returned then the command likely didn't return anything in the first place. i.e. delete
func ExecuteCommand(client kubernetes.Interface, command Command) (interface{}, error) {
	switch command.Verb {
	case "get":
		switch command.Resource {
		case "pod", "pods":
			if command.Identifier != "" {
				return k8s.GetPod(client, command.Namespace, command.Identifier)
			} else {
				return k8s.GetPods(client, command.Namespace)
			}
		default:
			return nil, errors.New(fmt.Sprintf("failed to execute command - unknown resource: %s", command.Resource))
		}

	case "delete":
		switch command.Resource {
		case "pod", "pods":
			if command.Identifier == "" {
				return nil, errors.New(fmt.Sprintf("attempted delete command execution without identifier specified: %v", command))
			}
			return nil, k8s.DeletePod(client, command.Namespace, command.Identifier)
		default:
			return nil, errors.New(fmt.Sprintf("failed to execute command - unknown resource: %s", command.Resource))
		}

	default:
		return nil, errors.New(fmt.Sprintf("failed to execute command - unknown verb: %s", command.Verb))
	}

}

// parseAndValidateCommandFromString parses a given string and returns the corresponding Command struct if valid
func parseAndValidateCommandFromString(command string) (Command, error) {

	commandArr := strings.Split(command, " ")

	withIdentifier := false
	switch len(commandArr) {
	case 3:
		withIdentifier = false
	case 4:
		withIdentifier = true
	default:
		return Command{}, errors.New("unable to parse Command")
	}

	verb, err := checkPermission(commandArr, verbPermissionIndex, permissions.Verbs)
	if err != nil {
		return Command{}, err
	}
	resource, err := checkPermission(commandArr, resourcePermissionIndex, permissions.Resources)
	if err != nil {
		return Command{}, err
	}
	namespace, err := checkPermission(commandArr, namespacePermissionIndex, permissions.Namespaces)
	if err != nil {
		return Command{}, err
	}
	identifier := ""
	if withIdentifier {
		identifier = commandArr[3]
	}

	return Command{
		Verb:       verb,
		Resource:   resource,
		Namespace:  namespace,
		Identifier: identifier,
	}, nil
}

func checkPermission(commandArr []string, idx int, allowedValues []string) (string, error) {
	value := commandArr[idx]
	valueSupported := util.StringInSliceIgnoreCase(value, allowedValues)
	if !valueSupported {
		return "", errors.New(fmt.Sprintf("permission error - unsupported value: %s at index %d", value, idx))
	}
	return value, nil
}

// renderteamsPodCard will render an adaptive teams card for each pod found
func renderTeamsPodCard(podList []v1.Pod) ([]byte, error) {
	tmpl := template.New("teams-adaptive-card-pod-list.tmpl")
	tmpl, err := tmpl.Funcs(templateFns).Parse(teamsAdaptiveCardPodListTmpl)
	if err != nil {
		return nil, err
	}

	var tmplData bytes.Buffer
	err = tmpl.Execute(&tmplData, podList)
	if err != nil {
		return nil, err
	}

	jsonTeamsCard, err := json.MarshalIndent(tmplData.String(), "", " ")
	if err != nil {
		return nil, err
	}
	return jsonTeamsCard, nil
}
