[![kaindlglobalnetwork.de](https://raw.githubusercontent.com/kgncloud/docker-template/main/KAINDL_LogoDesign_B_orange_rgb.png)](https://kaindlglobalnetwork.de)

# mautrix-whatsapp (Kaindl Network edition)

A Matrix-WhatsApp puppeting bridge based on [whatsmeow](https://github.com/tulir/whatsmeow).

This repository is a fork of [mautrix/whatsapp](https://github.com/mautrix/whatsapp).
It carries the full upstream source and history plus the following additions:

* **Call bridging** – incoming WhatsApp calls show up in the Matrix room, the notice is
  updated live (answered on another device, missed, declined, ended with duration) and a
  ringing call can be declined from Matrix with `!wa decline-call`.
  Audio/video bridging is planned, see [docs/calls.md](docs/calls.md) for status and design.
* **Container image built from source** with a built-in healthcheck, current security
  updates and without a package manager at runtime.

## Documentation
All setup and usage instructions of the upstream bridge apply, see [docs.mau.fi]:

[docs.mau.fi]: https://docs.mau.fi/bridges/go/whatsapp/index.html

* [Bridge setup](https://docs.mau.fi/bridges/go/setup.html?bridge=whatsapp)
  (or [with Docker](https://docs.mau.fi/bridges/general/docker-setup.html?bridge=whatsapp))
* Basic usage: [Authentication](https://docs.mau.fi/bridges/go/whatsapp/authentication.html)
* [ROADMAP.md](ROADMAP.md) gives an overview of what is supported.

## 🐳 Docker

```bash
docker run -d --restart=unless-stopped -v ./data:/data ghcr.io/kaindlnetwork/mautrix-whatsapp:main
```

On first start the container writes `/data/config.yaml` and exits. Edit it, start the
container again to generate `/data/registration.yaml`, register that with your homeserver
and start the container a third time.

### Build the image locally

```bash
git clone https://github.com/kaindlnetwork/mautrix-whatsapp
cd mautrix-whatsapp
docker build -t mautrix-whatsapp .
```

### Volumes

| Volume | Description |
| :----: | --- |
| `/data` | Config, registration file and (if configured) the SQLite database |

### Environment variables

| Variable | Default | Description |
| :----: | --- | --- |
| `UID` / `GID` | `1337` | User and group the bridge runs as |
| `BRIDGE_PORT` | `29318` | Port used by the healthcheck. Must match `appservice.port` in `config.yaml` |

### 🩺 Healthcheck

The image checks `http://localhost:$BRIDGE_PORT/_matrix/mau/live` every 30 seconds.

## Keeping up to date with upstream

```bash
git remote add upstream https://github.com/mautrix/whatsapp
git fetch upstream
git merge upstream/main
```

## Licence

The bridge is licensed under the GNU Affero General Public License v3.0 or later, see [LICENSE](LICENSE).

## Discussion
Upstream Matrix room: [#whatsapp:maunium.net](https://matrix.to/#/#whatsapp:maunium.net)
