Installation
============

Prerequisites
-------------
In order to use the apt-transport-cloudflared package you must have
`cloudflared` installed. Follow the instructions
[here](https://developers.cloudflare.com/argo-tunnel/downloads/) to
install `cloudflared`.

Building and Installing
-----------------------
Before building `apt-transport-cloudflared`, please ensure that a go
development environment is set up. This can be done by following the
instructions provided [here](https://golang.org/doc/install). Once that
is done, use `go get` to download and install the
`apt-transport-cloudflared` binary to `${GOPATH}/bin`, then move the
binary to the correct location.

User Configuration
==================
Once the `apt-transport-cloudflared` binary is installed, you should be
able to use it by adding the access protected endpoint to your apt
sources with the scheme changed to `cfd+https://`. I.e.

```
deb [arch=amd64] cfd+https://my.apt-repo.org/v2/stretch stable common
```

Using Apt-Transport-Cloudflared
===============================
If everything is set up correctly, using the method should work
seamlessly. If the access token needs to be updated, a browser window
should open automatically and redirect to the root of your apt
repository. If this does not happen, the auth URL will be printed in
the apt output like so:

```
$ sudo apt update
AuthURL: https://my.apt-repo.org/cdn-cgi/access/cli?redirect_url=...
```

To avoid having this happen, you can log-in with `cloudflared` prior to
running the `apt update` or `apt install` commands, e.g

```
$ cloudflared access login https://my.apt-repo.org
$ sudo apt update && sudo apt install ${PACKAGES}
```

Service Tokens
==============
As an extension, the apt-transport-cloudflared package supports using
service tokens to authenticate with Access. Service token support is
implemented through two headers, one for the client ID and one for the
client secret. These values are passed to the service using the
`CF-Access-Client-ID` and `CF-Acess-Client-Secret` headers,
respectively.

Service tokens are accessed from `${HOME}/.cloudflared/servicetokens/`
with a filename corresponding to the root URL of the repository, and
are expected to have the following contents:

```
${CLIENT_ID}
${CLIENT_SECRET}
```

As an example, given a repository at `access.widgetcorp.tech` which
uses Access, in order to use a service token you would add a file to
`${HOME}/.cloudflared/servicetokens/access.widgetcorp.tech-Service-Token`
with the following contents:

```
bd2744144725d2651d39363df6807599.access.widgetcorp.tech
3e2c2ad371b00777a443f0c639c1e03687e4fcf73e0c3371cb1cbd6124b123fdef782a
```

Since the service tokens are already valid as is, using them does not
require `cloudflared`.
