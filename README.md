[![kaindlglobalnetwork.de](https://raw.githubusercontent.com/kgncloud/docker-template/main/KAINDL_LogoDesign_B_orange_rgb.png)](https://kaindlglobalnetwork.de)

# mautrix-whatsapp

The [Kaindl Network](https://kaindlglobalnetwork.de) team brings you another container release featuring:

* Regular application & security updates
* Built-in Healthcheck
* bringing a container as close as possible to linuxserver.io releases
* We track the upstream source changes and publishes fast new versions of this image.
* Bringing you Linuxserver.io Image Quality
* Using Alpine Base Image

## 🔧 How to Install

### 🐳 Docker

#### Docker CLI
```bash
docker run -d --restart=unless-failed -v ./data:/data ghcr.io/kgncloud/mautrix-whatsapp:main
```
#### Docker-Compose

### Build the Image locally

```bash
git clone https://github.com/kgncloud/mautrix-whatsapp/
cd mautrix-whatsapp
docker build Dockerfile
```

## What is needed to launch this Container

- Matrix Server with access to register bots
- Whatsapp Account
- Server(Linux preferred) to launch this Container 

## How does this Container work



## Version Tags

This image provides various versions that are available via tags. `main` tag usually provides the latest stable version. Others are considered under development and caution must be exercised when using them.

| Tag | Description |
| :----: | --- |
| main | Latest Release Version of this Container |

## Volumes

| Volumes | Description |
| :----: | --- |
| data | Here are the Registration Files, Bride Configuration Files (and SQLite File if configured) |

## 🩺 Healthcheck

This image has already a healtcheck included :D

## System Requirements

This Image was test for the use of one person

| Ressources | Description |
| :----: | --- |
| CPU | Low usage maximum 2 Threads needed |
| RAM | 1GB                                |
| Disk| No prediction possible, depends on user|

## Enviromental Variables
The Deveolper of the software generates a Configuration and a Registration File. We want to make your setup even easier by giving you the choise to input the configuration via Enviromental Variables. This would also allow an easier launch on Docker Swarm withut the need to expose your Filesystem.

| Variable | Description| Required Variable? | Standard Variable |
| :----: | --- | --- | --- |
| homeserver_adress | The address that this appservice can use to connect to the homeserver. | yes | https://example.com
| homeserver_domain | The domain of the homeserver (for MXIDs, etc).                         | yes | example.com
| asmux             | Is the homeserver actually mautrix-asmux?                              | no  | false
| status_endpoint   | The URL to push real-time bridge status to. If set, the bridge will make POST requests to this URL whenever a user's whatsapp connection state changes. The bridge will use the appservice as_token to authorize requests. | no | null
| message_send_checkpoint_endpoint | Endpoint for reporting per-message status. | no | null
| async_media | Does the homeserver support https://github.com/matrix-org/matrix-spec-proposals/pull/2246? | no | false
| bridge_adress | The address that the homeserver can use to connect to this appservice. | yes | http://localhost:80
| bridge_hostname | The hostname and port where this appservice should listen. | yes | 0.0.0.0
| bridge_port | The hostname and port where this appservice should listen.     | yes | 80
| database_type | The database type. "sqlite3" and "postgres" are supported.   | yes | sqlite3
| database_uri | | yes | mautrix-whatsapp.db
| database_max_open_conns | Maximum number of connections. Mostly relevant for Postgres. | no | 20
| database_max_idle_conns | Maximum number of connections. Mostly relevant for Postgres. | no | 2
| database_max_conn_idle_time | Maximum connection idle time and lifetime before they're closed. Disabled if null. | no | null
| database_max_conn_lifetime | Maximum connection idle time and lifetime before they're closed. Disabled if null. | no | null





## Licence

Our Images are always having the Unlicence if possible. The Developer of this Image did use the GNU Affero General Public License v3.0 . This Licence gives us the requirement to passthrough the same Licence to be compliant.

### Permissions

- Commercial use
- Modification
- Distribution
- Patent use
- Private use

### Limitations

- Liability
- Warranty

### Conditions

- License and copyright notice
- State changes
- Disclose source
- Network use is distribution
- Same license

docker pull ghcr.io/kgncloud/mautrix-whatsapp:main
