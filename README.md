# tailsocks

## Overview

A simple SOCKS5/HTTP proxy server for tailnet without installing tailscale client.

## Usage

Generate a new auth key in the [admin panel](https://login.tailscale.com/admin/settings/keys)
and set the TS_AUTHKEY environment variable.

## fly.io

Deploy to fly.io.

```sh
flyctl secrets set TS_AUTHKEY=XXXXXXXX
flyctl deploy
```