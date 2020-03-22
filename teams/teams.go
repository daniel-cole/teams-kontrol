package teams

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daniel-cole/teams-kontrol/middleware"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const KontrolSharedSecretEnvKey = "TEAMS_KONTROL_SHARED_SECRET"

type Request struct {
	Type           string    `json:"type"`
	ID             string    `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	LocalTimestamp time.Time `json:"localTimestamp"`
	ServiceURL     string    `json:"serviceUrl"`
	ChannelID      string    `json:"channelId"`
	From           struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AadObjectID string `json:"aadObjectId"`
	} `json:"from"`
	Conversation struct {
		IsGroup          bool        `json:"isGroup"`
		ID               string      `json:"id"`
		Name             interface{} `json:"name"`
		ConversationType string      `json:"conversationType"`
		TenantID         string      `json:"tenantId"`
	} `json:"conversation"`
	Recipient        interface{}   `json:"recipient"`
	TextFormat       string        `json:"textFormat"`
	AttachmentLayout interface{}   `json:"attachmentLayout"`
	MembersAdded     []interface{} `json:"membersAdded"`
	MembersRemoved   []interface{} `json:"membersRemoved"`
	TopicName        interface{}   `json:"topicName"`
	HistoryDisclosed interface{}   `json:"historyDisclosed"`
	Locale           string        `json:"locale"`
	Text             string        `json:"text"`
	Speak            interface{}   `json:"speak"`
	InputHint        interface{}   `json:"inputHint"`
	Summary          interface{}   `json:"summary"`
	SuggestedActions interface{}   `json:"suggestedActions"`
	Attachments      []struct {
		ContentType  string      `json:"contentType"`
		ContentURL   interface{} `json:"contentUrl"`
		Content      string      `json:"content"`
		Name         interface{} `json:"name"`
		ThumbnailURL interface{} `json:"thumbnailUrl"`
	} `json:"attachments"`
	Entities []struct {
		Type     string `json:"type"`
		Locale   string `json:"locale"`
		Country  string `json:"country"`
		Platform string `json:"platform"`
	} `json:"entities"`
	ChannelData struct {
		TeamsChannelID string `json:"teamsChannelId"`
		TeamsTeamID    string `json:"teamsTeamId"`
		Channel        struct {
			ID string `json:"id"`
		} `json:"channel"`
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
		Tenant struct {
			ID string `json:"id"`
		} `json:"tenant"`
	} `json:"channelData"`
	Action    interface{} `json:"action"`
	ReplyToID interface{} `json:"replyToId"`
	Value     interface{} `json:"value"`
	Name      interface{} `json:"name"`
	RelatesTo interface{} `json:"relatesTo"`
	Code      interface{} `json:"code"`
}

type Response struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var secret string

// Init will ensure that there's a valid shared secret. The program will crash if it is not specified.
func Init() {
	if secret = os.Getenv(KontrolSharedSecretEnvKey); secret == "" {
		logrus.Fatalf("Exiting. Please specify a shared secret with: %s", KontrolSharedSecretEnvKey)
	}
}

func parseTeamsRequestText(text string) string {
	r := regexp.MustCompile(`^.*</at> (.*)\n$`)
	match := r.FindStringSubmatch(text)
	if match != nil {
		return match[1]
	}
	return ""
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		middleware.LogWithContext(ctx).Errorf("failed to decode request body from teams: %v", err)
		teamsResponse := &Response{
			Type: "message",
			Text: "failed to parse payload from teams",
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(teamsResponse)
		return
	}

	middleware.LogWithContext(ctx).Infof("Received request from %s", request.From.Name)
	// here we should parse the text and then attempt to execute the command if it's valid
	//parsedText := parseTeamsRequestText(request.Text)

	msg := fmt.Sprintf("%s - that command is not available. Please specify a valid command.", request.From.Name)
	teamsResponse := &Response{
		Type: "message",
		Text: msg,
	}

	middleware.LogWithContext(ctx).Info("Finished processing request")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(teamsResponse)
}

// AuthHandler provides http middleware to authenticate the outgoing teams request with HMAC
func AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		auth := r.Header.Get("Authorization")
		if auth == "" {
			middleware.LogWithContext(ctx).Errorf("no auth header set from client")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// read the body as it's used to compute the HMAC
		body, err := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			middleware.LogWithContext(ctx).Errorf("failed to parse body from client")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		middleware.LogWithContext(ctx).Debugf("Payload from client: %s", string(body))

		expectedMAC := strings.TrimPrefix(auth, "HMAC ")
		verifiedMAC, err := verifyMAC(secret, expectedMAC, string(body))
		if err != nil {
			middleware.LogWithContext(ctx).Errorf("Failed to verify MAC: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if verifiedMAC { // user authenticated
			// set the request body for the next request as we've already read it
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		middleware.LogWithContext(ctx).Infof("Attempted unauthorized access to protected endpoint")
	})
}

// verifyMAC checks if the MAC computed with the base64 encoded secret
// matches the expectedMAC given the payload. expectedMAC is also expected to be encoded in base64
func verifyMAC(secret string, expectedMAC string, payload string) (bool, error) {
	secretBytes, err := base64.StdEncoding.DecodeString(secret)

	if err != nil {
		return false, errors.New("failed to decode secret")
	}

	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(payload))
	computedMACBytes := h.Sum(nil)
	expectedMACBytes, err := base64.StdEncoding.DecodeString(expectedMAC)
	if err != nil {
		return false, errors.New("failed to decode expected MAC")
	}

	return hmac.Equal(computedMACBytes, expectedMACBytes), nil
}
