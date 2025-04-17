# Tailscale SOCKS5/HTTP proxy server

## Overview

This project starts a SOCKS5 (port 1080) and HTTP (port 3128) proxy server on a Tailscale client.

## Usage

- Generate a new auth key from the [Tailscale admin panel](https://login.tailscale.com/admin/settings/keys)
- Set the TS_AUTHKEY environment variable with the generated key.

## Deploy to fly.io

You can deploy this proxy server to [fly.io](https://fly.io) using the following commands:

```sh
flyctl secrets set TS_AUTHKEY=XXXXXXXX
flyctl deploy
```