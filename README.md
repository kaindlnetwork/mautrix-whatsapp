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

docker pull ghcr.io/kgncloud/mautrix-whatsapp:main
