# teams-kontrol

:warning: This project is currently under development. In it's current state it should only be used as a reference :warning:

> Your scientists were so preoccupied with whether or not they could, they didn't stop to think if they should

This tool provides an interface for interacting with Kubernetes via Microsoft Teams outgoing webhooks.
For now I'm assuming whoever is interested in this tool has a grasp on Kubernetes.

It's been created in a way that can be easily extended to support other outgoing webhooks from different providers.
I don't think I'll add support for other providers because this is a small weekend project.

# Deployment
After building the binary you should deploy it to Kubernetes using the provided example manifest as a reference:
[deployment.yml.example](deployment.yml)

For a full list of environment variables please see the [envrc.example](envrc.example)

## Permissions
There's two levels of permissions that you'll need to define:
1. Kubernetes RBAC
2. Application permissions

### Kubernetes RBAC
This is standard and you can find an example in the manifest [deployment.yml.example](deployment.yml)

### Application permissions
This is a second layer of permissions that the application will read in at runtime.

The file is located by specifying the environment variable `TEAMS_KONTROL_PERMISSION_FILE`. This defaults to `permissions.yml`
In the [example](deployment.yml) deployment manifest this is file is created as a config map and mounted to the pod.

The structure of the permissions file is similar to k8s RBAC and is as follows:

```
verbs:
  - "get"
  - "delete"
namespaces:
  - "default"
resources:
  - "pods"
```
See: [permissions.yml.example](permissions.yml.example)

# How it works

After you've created an outgoing webhook in teams and pointed it to your deployment you can execute commands by running:
`@<your outgoing webhook> get pods nginx`

# Teams cards

Example card generated from a command. i.e. `get pods default`
```
{
  "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
  "type": "AdaptiveCard",
  "version": "1.0",
  "body": [
    {
      "type": "TextBlock",
      "id": "d255f1e1-8d61-418f-99d3-13374e0fe7fe",
      "text": "Pod Detail",
      "wrap": true,
      "size": "Large",
      "weight": "Bolder",
      "color": "Accent",
      "horizontalAlignment": "Center"
    },
    {
      "type": "Container",
      "id": " 720e735f-384b-4d1e-966c-7e44a9d6287d",
      "padding": "None",
      "items": [
        {
          "type": "FactSet",
          "id": "0ed931cd-74b8-4775-8905-0347ee778410",
          "facts": [
            {
              "title": "Name",
              "value": "nginx-9ffc7d87b-fw2vd"
            },
            {
              "title": "Age",
              "value": "2020-03-18 02:08:55 +0000 UTC"
            },
            {
              "title": "Status",
              "value": "Running"
            },
            {
              "title": "Namespace",
              "value": "default"
            }
          ]
        }
      ],
      "style": "emphasis"
    }
  ],
  "padding": "None"
}

```

# Development

Starting up a kind cluster and loading in the image is the probably the easier way to start testing quickly.

You can set the following environment variable to execute commands directly without using HMAC auth via the teams handler:

`TEAMS_KONTROL_INSECURE_COMMANDS=TRUE`

Then you can issue curl commands to the endpoint. i.e. `curl http://<host>/command -d "get pods default"`

## Adaptive Teams Cards
To create and edit samples before creating templates you can use: https://amdesigner.azurewebsites.net/
