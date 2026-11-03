# Installer: Pi disk image

[`src/installer/custompios/`](../src/installer/custompios/) builds the SD card image for the Pi Zero W with [CustomPiOS](https://github.com/guysoft/CustomPiOS) 2.x. It starts from Raspberry Pi OS Lite 32-bit (armhf, Trixie) and adds the app, its systemd service and the Wi-Fi access point.

## Build the image

Requirements: Docker (Docker Desktop on macOS, Apple Silicon or Intel; or Docker on Linux), plus Go and Node as for `make`.

```bash
make image
```

This runs the tests, cross-compiles the app for the Pi and runs [`build-image.sh`](../src/installer/custompios/build-image.sh), which:

1. downloads the base image pinned in [`config`](../src/installer/custompios/config) (`TIMELAPSEPI_BASE_IMAGE_URL`) into `image/` and checks it against `TIMELAPSEPI_BASE_IMAGE_SHA256`. The download happens once; later builds reuse it.
2. checks that Docker can run 32-bit ARM programs. The image's programs run under qemu, which the host kernel has to know about (binfmt_misc). Docker Desktop registers it; on a Linux host without it, run `docker run --privileged --rm tonistiigi/binfmt --install arm` once per boot (CI uses `docker/setup-qemu-action`).
3. runs CustomPiOS in its pinned Docker image (`ghcr.io/guysoft/custompios@sha256:…`) with `--privileged`, because it loop-mounts the image's partitions and chroots into them. Ctrl-C stops the build and unmounts the image. Afterwards the script detaches the loop devices the build left behind and makes `workspace/` owned by you (the container runs as root).
4. writes `src/installer/custompios/workspace/<base image date>-timelapsepi-trixie-armhf-lite-<version>.zip`, where the version comes from `git describe`.

Write the image to an SD card with [Raspberry Pi Imager](https://www.raspberrypi.com/software/) ("Use custom"), without its OS customisation, since the image already contains the user, SSH and Wi-Fi settings.

To move to a newer Raspberry Pi OS release, change the URL and SHA-256 in `config` (the SHA-256 is published next to each image as `<image>.img.xz.sha256`). To update CustomPiOS, change `CUSTOMPIOS_IMAGE` in `build-image.sh` to the digest of a newer `ghcr.io/guysoft/custompios` image.

Local settings that should not be committed go into `src/installer/custompios/config.local` (sourced last; gitignored).

### In CI

[`.github/workflows/image.yml`](../.github/workflows/image.yml) runs `make image` on an `ubuntu-24.04-arm` runner when started by hand ("Run workflow") and for every pushed tag. The zipped image is a workflow artifact, kept for one day (it is larger than the 500 MB of artifact storage on GitHub Free). For a tag, a second job, the only one with write access to the repository, attaches it to the GitHub release of that tag (creating the release if needed).

## How the app gets into the image

`make pi` cross-compiles the app for the Pi Zero W into `modules/timelapsepi/filesystem/root/opt/timelapse/timelapse` (gitignored). The image build copies it to `/opt/timelapse/timelapse`, and `timelapse.service` runs it from there. The image therefore contains exactly the app of your checkout; nothing is cloned from GitHub at build time, and the image has no git checkout of this repo.

- The service runs as `pi` with `WorkingDirectory=/home/pi`, so `timelapse.log` (rotated at 10 MB to `timelapse.log.1`) and `timelapse-settings.db` live in `/home/pi`. It restarts automatically (`Restart=always`) after crashes and failed starts.
- `/etc/sudoers.d/010_timelapse` lets `pi` run exactly `/sbin/shutdown -h now` and `/sbin/shutdown -r now`, which the web UI uses to shut down and restart the Pi.
- To update the app on a running Pi without reflashing, build it with `make pi`, connect to the Pi's access point and run:

  ```bash
  scp src/installer/custompios/modules/timelapsepi/filesystem/root/opt/timelapse/timelapse pi@192.168.50.1:/tmp/
  ssh -t pi@192.168.50.1 "sudo install -m 755 /tmp/timelapse /opt/timelapse/timelapse && sudo systemctl restart timelapse"
  ```

### Devices with older images

Images up to release 0.5.0 (and those built from the main branch before the app was baked in) ran the app from a git clone of this repository, partly as root and with the legacy camera stack configured. The repository no longer contains a prebuilt binary, so `git pull` on such a device removes the binary its service runs. Reflash these devices with a new image. Download the photos first (web UI, or `/archive/zip`); the settings have to be entered again.

## Real-time clock

The Pi Zero W has no clock of its own, and as an access point it never reaches a time server. After a power cut its clock restarts from the time it last saved, so photo names and hour-aligned captures are off. A battery-backed RTC such as the [Adafruit PiRTC (PCF8523)](https://www.adafruit.com/product/3386) fixes this (wiring: [docs/hardware.md](hardware.md)).

The image always contains the parts that read the RTC at boot: `util-linux-extra` (`hwclock`), `/etc/udev/rules.d/90-timelapse-rtc.rules` and `timelapse-rtc.service`, plus `i2c-tools` and `/etc/modules-load.d/i2c-dev.conf` for checking the chip. When the RTC driver loads, udev starts the service, which sets the system clock from the RTC. Without an RTC nothing runs. If the app started first, its scheduler picks up the new time within a minute.

Enabling the driver:

- On an SD card: open `config.txt` on the boot partition from any computer and uncomment `#dtoverlay=i2c-rtc,pcf8523` at the end (for a DS3231 or DS1307, use `ds3231` or `ds1307`).
- When building: set `TIMELAPSEPI_RTC=pcf8523` in `src/installer/custompios/config.local` and the line is active in the image.

The RTC has to be set once. Connect to the Pi's access point and run from your computer:

```bash
ssh -t pi@192.168.50.1 "sudo date -u -s @$(date +%s) && sudo hwclock --systohc --utc"
```

(`-t` lets `sudo` ask for the password.) Check it with `sudo hwclock -r` on the Pi. `sudo i2cdetect -y 1` (the image loads `i2c-dev` for it) shows `UU` at address 68 once the PCF8523 driver uses the chip.
