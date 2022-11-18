# tailsocks5

## Overview

A simple socks5 server for tailnet.

## Usage

Generate a new auth key in the [admin panel](https://login.tailscale.com/admin/settings/keys)
and set the TS_AUTHKEY environment variable.

## fly.io

Deploy to fly.io.

```sh
flyctl secrets set TS_AUTHKEY=XXXXXXXX
flyctl deploy
```