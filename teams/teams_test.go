package teams

import (
	"bytes"
	"encoding/json"
	"github.com/daniel-cole/teams-kontrol/util"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testRequest = `{
  "type": "message",
  "id": "1",
  "timestamp": "2020-03-18T07:17:22.7590435Z",
  "localTimestamp": "2020-03-18T17:17:22.7590435+10:00",
  "serviceUrl": "https://trafficmanager.net/",
  "channelId": "msteams",
  "from": {
    "id": "id",
    "name": "Daniel Cole",
    "aadObjectId": "id"
  },
  "conversation": {
    "isGroup": true,
    "id": "id",
    "name": null,
    "conversationType": "channel",
    "tenantId": "id"
  },
  "recipient": null,
  "textFormat": "plain",
  "attachmentLayout": null,
  "membersAdded": [],
  "membersRemoved": [],
  "topicName": null,
  "historyDisclosed": null,
  "locale": "en-US",
  "text": "<at>teams-kontrol</at> debug last time\n",
  "speak": null,
  "inputHint": null,
  "summary": null,
  "suggestedActions": null,
  "attachments": [
    {
      "contentType": "text/html",
      "contentUrl": null,
      "content": "<div><div><span itemscope=\"\" itemtype=\"http://schema.skype.com/Mention\" itemid=\"0\">teams-kontrol</span> debug last time</div>\n</div>",
      "name": null,
      "thumbnailUrl": null
    }
  ],
  "entities": [
    {
      "type": "clientInfo",
      "locale": "en-US",
      "country": "US",
      "platform": "Web"
    }
  ],
  "channelData": {
    "teamsChannelId": "id",
    "teamsTeamId": "id",
    "channel": {
      "id": "id"
    },
    "team": {
      "id": "id"
    },
    "tenant": {
      "id": "id"
    }
  },
  "action": null,
  "replyToId": null,
  "value": null,
  "name": null,
  "relatesTo": null,
  "code": null
}`

func TestMain(m *testing.M) {
	teamsKontrolSharedSecret := "c2VjcmV0Cg==" // secret is "secret"
	err := util.AttemptSetEnv(KontrolSharedSecretEnvKey, teamsKontrolSharedSecret)
	if err != nil {
		log.Fatalf("Failed to set %s", KontrolSharedSecretEnvKey)
	}
	Init()
	os.Exit(m.Run())
}

func TestParseTeamsRequestText(t *testing.T) {
	var request Request
	err := json.Unmarshal([]byte(testRequest), &request)
	if err != nil {
		t.Fatal("Failed to unmarshal JSON to request")
	}
	text := request.Text
	parsedText := parseTeamsRequestText(text)
	expectedParsedText := "debug last time"
	if parsedText != expectedParsedText {
		t.Fatalf("Parsed text does not match expected text, got %s, expected %s", parsedText, expectedParsedText)
	}
}

func TestHandleMessage(t *testing.T) {
	var request Request
	err := json.Unmarshal([]byte(testRequest), &request)
	if err != nil {
		t.Fatal("Failed to unmarshal JSON to request")
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatal("Failed to marshal JSON request")
	}

	req, err := http.NewRequest("POST", "/teams", bytes.NewBuffer(jsonRequest))

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-type", "Application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MessageHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expectedResponse := `{"type":"message","text":"Daniel Cole - that command is not available. Please specify a valid command."}` + "\n"
	response := rr.Body.String()
	if response != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v expected : %v", response, expectedResponse)
	}

}

func TestTeamsAuth(t *testing.T) {

	var request Request
	err := json.Unmarshal([]byte(testRequest), &request)
	if err != nil {
		t.Fatal("Failed to unmarshal JSON to request")
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatal("Failed to marshal JSON request")
	}

	req, err := http.NewRequest("POST", "/teams", bytes.NewBuffer(jsonRequest))

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-type", "Application/json")
	req.Header.Add("Authorization", "HMAC AUcyAKsiB2yCYuFhsz6O9qQ0gY+hQFL3IDxbTJhMWFY=")

	rr := httptest.NewRecorder()
	handler := AuthHandler(http.HandlerFunc(MessageHandler))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expectedResponse := `{"type":"message","text":"Daniel Cole - that command is not available. Please specify a valid command."}` + "\n"
	response := rr.Body.String()
	if response != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v expected : %v", response, expectedResponse)
	}

}

func TestVerifyMac(t *testing.T) {
	secret := "c2VjcmV0Cg==" // secret is "secret"
	expectedMAC := "hXdVxuUH16fgQm1bM9Y+EcBfGmTkcoVdmT9Om0HlLmA="
	payload := `{"hello":"from","hmac":"foryou"}`
	verified, err := verifyMAC(secret, expectedMAC, payload)

	if err != nil {
		t.Fatalf("error encountered when attempting to verify MAC: %v", err)
	}
	if !verified {
		t.Fatalf("expected MAC to be verified")
	}
}
