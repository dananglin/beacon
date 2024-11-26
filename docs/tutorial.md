<!--
SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>

SPDX-License-Identifier: CC-BY-4.0
-->

# Tutorial

This tutorial will go through the process of building and deploying an instance of Beacon.

In this example scenario, I have a domain and a personal static website at `https://dananglin.example` and I want
to deploy an instance of Beacon at `https://auth.dananglin.example` so that I can sign into client sites with my
domain with IndieAuth.

## Installation

## Requirements

### Build requirements

The following tools are needed to build the binary:

- **Git:** Required for cloning the repository. Please go [here](https://git-scm.com/downloads) for instructions on how to download Git for your repository.
- **Go:** A minimum version of Go 1.23.3 is required for installing Enbas. Please go [here](https://go.dev/dl/) to download the latest version.
- **Mage (optional):** The project includes mage targets for building the binary and docker image. The main advantage of using mage over just using `go build` is that the build information is built into the binary during compilation. You can visit the [Mage website](https://magefile.org/) for instructions on how to install Mage.

### Deployment requirements

- A (sub)domain that can be resolved to the IP address of your Beacon instance.
- A reverse proxy (e.g. [Caddy](https://caddyserver.com/)) that can perform TLS termination with a valid certificate.

This tutorial assumes that you have both the DNS record for the domain and the reverse proxy already set up.

## Build

### Build the binary

Clone the repository.
```bash
git clone https://github.com/dananglin/beacon.git
```

#### Build with mage

If you have mage installed you can build the binary with:
```bash
mage clean build
```

You can set the following environment variables when building with Mage:

| Environment variable | Description |
|----------------------|-------------|
| `BEACON_BUILD_REBUILD_ALL` | Set this to `"1"` to rebuild all packages even if they are already up-to-date.<br>(i.e. `go build -a ...`)|
| `BEACON_BUILD_VERBOSE` | Set this to `"1"` to enable verbose logging when building the binary.<br>(i.e. `go build -v ...`)|

#### Build with go

Otherwise you can build the binary with `go build`:
```bash
go build -a -ldflags="-s -w" -o ./__build/beacon .
```

This will build the binary to the `__build` directory.
You can install the binary to one of the directories in your `PATH`.

### Build the docker image

There is a mage target for building a docker image.
This builds both the binary and the docker image.
You can build the docker image with:
```
mage docker
```

You can set the following environment variables when building with Mage:

| Environment variable | Description |
|----------------------|-------------|
| `BEACON_DOCKER_IMAGE_NAME` | Use this to specify the name of the docker image. |
| `BEACON_DOCKERFILE` | Use this to specify a different Dockerfile.<br>By default the `Dockerfile` is used. |

If you don't wish to use mage then you can build the docker image using `docker build`.
Ensure that you've built the binary before building the docker image.
```
docker build -t localhost/beacon:latest .
```

## Configure

### Create your configuration file

You'll need to create a JSON file to configure your Beacon instance.
You can copy the example configuration [here](../example/config.json) and edit for your setup.
Please refer to the [configuration reference](./configuration.md) for help with your configuration.

### Generate your JWT secret

You'll need to create a JWT secret for signing JWT access tokens.
This can be any random string; preferably 32 characters or more.
You can use the following example command to generate the secret:
```bash
openssl rand -base64 32
```

Once you've generated the secret add it to your configuration.

### Example configuration for this tutorial

```json
{
    "bindAddress": "0.0.0.0",
    "port": 8080,
    "domain": "auth.dananglin.example",
    "database": {
        "path": "./data/beacon.db"
    },
    "jwt": {
        "secret": "18ydyvTHmtdA/CMwyDGKjpm+j+oPgvnjQUO1qAkJc8w=",
        "cookieName": "beacon-5G5yyt9pfeJUUUeq"
    }
}
```

## Deploy

### Run the binary

Here's an example script to run the binary.

```bash
mkdir data
beacon serve --config config.json
```

### Run the docker image

```bash
mkdir data
docker run -d \
    -v ./config.json:/config.json \
    -v ./data:/data \
    -p 8080:8080 \
    --name beacon \
    --restart always \
    localhost/beacon:latest serve --config=/config.json
```

Once you have Beacon running, ensure that your reverse proxy and DNS records are routing traffic to your VM or Docker container.

## Setup your profile

Open your favourite browser and go to the URL of your Beacon instance.
You'll be presented with a form to set up your profile.

### Profile ID and password

The `Profile ID` (a.k.a your Profile URL) is the URL of your domain or website. A valid profile ID must follow the following requirements:

- It must have either the `https` or `http` scheme.
- It must contain a path component (`/` is a valid path).
- It may contain a query string component (e.g. `https://website.example?userID=1001`).
- It must not contain a fragment component.
- It must not contain a username or password component.
- It must not contain a port.
- It must not be a IPv4 or IPv6 address.

You can view the requirements in the [User Profile URL](https://indieauth.spec.indieweb.org/#user-profile-url) section of the IndieAuth specification as well as examples of valid and invalid URLs.

Before creating your profile, Beacon validates the value of your profile ID against the above requirements and [canonicalizes](https://indieauth.spec.indieweb.org/#url-canonicalization) it before setting it as your profile ID.
For example `bobsmith.example` will be canonicalized to `https://bobsmith.example/`.

Next, enter and confirm your password.

### The profile information fields

The remaining fields are for your profile information which an IndieAuth client may request to see when you sign into their service.
These are not required for this setup and you can update them after you've created your profile.

- The `Profile display name` is the name you expect the client to use as your display name.
- The `Profile URL` is the URL of your website. This can be the same as your profile ID or even a URL of a different website.
- The `Profile photo URL` is the URL of an image that you wish the client to use as your profile image.
- The `Profile email` is the email address that you wish to provide to the client if the client requests to view your email address.

See the [Profile Information](https://indieauth.spec.indieweb.org/#profile-information) section of the IndieAuth specification to see more information about the profile information.

<div style="text-align:center">
<img src="./assets/images/setup.png"
     alt="Example setup form"
     width="495"
     height="540">
</div>

Once you have entered your details click the `Create profile` button.
You'll be redirected to the login page where you can sign into Beacon.

## Update your profile

Once you've signed into Beacon you'll be presented with a page with your profile information which you can
update at any time.

<div style="text-align:center">
<img src="./assets/images/profile.png"
     alt="Example setup form"
     width="554"
     height="409">
</div>

## Update your domain or website

Now that you've set up your profile in Beacon, you'll need to update your website so that IndieAuth clients can
discover your instance's authorization and token endpoint. According to the IndieAuth specification, clients
should discover these using the `indieauth-metadata` endpoint however older clients may search for the
`authorization_endpoint` and `token_endpoint` URLs in your site's HTML document.

See the [Discovery by Clients](https://indieauth.spec.indieweb.org/#discovery-by-clients) section of the specification for more information on how clients discover the endpoints.

The following sections will describe how you can update your website to configure all endpoints so that all
clients are supported.

### Set the indieauth-metadata HTTP header

Your instance's `indieauth-metadata` endpoint can be found at `https://<YOUR_BEACON_URL>/.well-known/oauth-authorization-server`.
To confirm enter the URL in your browser or HTTP client.
You should receive a JSON document that includes both the authorization and token endpoints.

```bash
curl -s https://auth.dananglin.example/.well-known/oauth-authorization-server | jq .
```

```json
{
  "issuer": "https://auth.dananglin.example/",
  "authorization_endpoint": "https://auth.dananglin.example/indieauth/authorize",
  "token_endpoint": "https://auth.dananglin.example/indieauth/token",
  "service_documentation": "https://indieauth.spec.indieweb.org",
  "code_challenge_methods_supported": [
    "S256"
  ],
  "grant_types_supported": [
    "authorization_code"
  ],
  "response_types_supported": [
    "code"
  ],
  "scopes_supported": [
    "profile",
    "email"
  ],
  "authorization_response_iss_parameter_supported": true
}
```

_TODO_: Finish section

### Set the authorization_endpoint and token_endpoint link tag

_TODO_

## Sign into an IndieAuth client

_TODO_
