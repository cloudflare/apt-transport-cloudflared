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
