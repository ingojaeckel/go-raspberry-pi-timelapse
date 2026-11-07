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

- The service runs as `pi` with `WorkingDirectory=/home/pi`, so `timelapse.log` (rotated at 10 MB to `timelapse.log.1`), `timelapse-settings.db` and `timelapse.running` (see [Health log](#health-log)) live in `/home/pi`. It restarts automatically (`Restart=always`) after crashes and failed starts.
- The service has the capability to set the system clock (`AmbientCapabilities=CAP_SYS_TIME`), for the web UI's clock button (see [Real-time clock](#real-time-clock)). Programs the app starts (`rpicam-still`, `vcgencmd`) inherit it.
- `/etc/sudoers.d/010_timelapse` lets `pi` run exactly `/sbin/shutdown -h now` and `/sbin/shutdown -r now`, which the web UI uses to shut down and restart the Pi, and `/usr/sbin/hwclock --rtc=/dev/rtc0 --systohc --utc`, which copies the clock set in the web UI to a clock module.
- To update the app on a running Pi without reflashing, build it with `make pi`, connect to the Pi's access point and run:

  ```bash
  scp src/installer/custompios/modules/timelapsepi/filesystem/root/opt/timelapse/timelapse pi@192.168.50.1:/tmp/
  ssh -t pi@192.168.50.1 "sudo install -m 755 /tmp/timelapse /opt/timelapse/timelapse && sudo systemctl restart timelapse"
  ```

### Devices with older images

Images up to release 0.5.0 (and those built from the main branch before the app was baked in) ran the app from a git clone of this repository, partly as root and with the legacy camera stack configured. The repository no longer contains a prebuilt binary, so `git pull` on such a device removes the binary its service runs. Reflash these devices with a new image. Download the photos first (web UI, or `/archive/zip`); the settings have to be entered again.

## Real-time clock

The Pi Zero W has no clock of its own, and as an access point it never reaches a time server. After a power cut its clock restarts from the time it last saved, so photo names and hour-aligned captures are off by however long the power was out.

Without extra hardware, set the clock from the web UI: when the camera's clock differs from your phone's by more than a minute, Status shows a warning with a "Set the camera's clock to this device's time" button (`POST /admin/time` with `{"UnixMillis": …}`). The change is logged. With a clock module, the button also writes the new time to it.

A battery-backed RTC such as the [Adafruit PiRTC (PCF8523)](https://www.adafruit.com/product/3386) fixes this (wiring: [docs/hardware.md](hardware.md)).

The image always contains the parts that read the RTC at boot: `util-linux-extra` (`hwclock`), `/etc/udev/rules.d/90-timelapse-rtc.rules` and `timelapse-rtc.service`, plus `i2c-tools` and `/etc/modules-load.d/i2c-dev.conf` for checking the chip. When the RTC driver loads, udev starts the service, which sets the system clock from the RTC. Without an RTC nothing runs. If the app started first, its scheduler picks up the new time within a minute.

Enabling the driver:

- On an SD card: open `config.txt` on the boot partition from any computer and uncomment `#dtoverlay=i2c-rtc,pcf8523` at the end (for a DS3231 or DS1307, use `ds3231` or `ds1307`).
- When building: set `TIMELAPSEPI_RTC=pcf8523` in `src/installer/custompios/config.local` and the line is active in the image.

The RTC has to be set once: with the web UI's clock button, or by connecting to the Pi's access point and running from your computer:

```bash
ssh -t pi@192.168.50.1 "sudo date -u -s @$(date +%s) && sudo hwclock --systohc --utc"
```

(`-t` lets `sudo` ask for the password.) Check it with `sudo hwclock -r` on the Pi. `sudo i2cdetect -y 1` (the image loads `i2c-dev` for it) shows `UU` at address 68 once the PCF8523 driver uses the chip.

## Time zone

The image sets the Pi's time zone to UTC (`TIMELAPSEPI_TIMEZONE=Etc/UTC` in the module's `config`; the base image uses `Europe/London`). Photo file names use it, and captures lined up with the clock follow its hours. Without daylight saving time no hour is skipped or repeated twice a year. The web UI shows times in the browser's time zone either way. For local file names, set e.g. `TIMELAPSEPI_TIMEZONE=Europe/Berlin` in `config.local`, or run `sudo timedatectl set-timezone Europe/Berlin` on the Pi.

## Health log

For a camera nobody looks at for weeks, the app's log (`timelapse.log`, `/logs` in the web UI) keeps what happened:

- At start-up, how the previous run ended: stopped cleanly; the app ended without stopping (crash, or killed by the systemd watchdog) and was restarted; or the Pi restarted while the app was running (power cut, hardware watchdog, kernel crash). The app keeps `timelapse.running` with the boot ID while it runs to tell these apart.
- A `daily summary` line every 24 hours of uptime (not at midnight, since the clock may be wrong): photos captured and failed with the last error, photos deleted to free space, CPU temperature range, under-voltage or throttling since boot, number of photos, free space, uptime. It is logged as a warning when a capture failed or the power was bad.
- Under-voltage and throttling as soon as they are seen (checked every 10 minutes with `vcgencmd get_throttled`; the firmware remembers them since boot, so short dips are not missed). Status in the web UI shows them too. Under-voltage is the usual cause of SD card corruption and failed captures on battery or solar power.

## Full card

By default the camera deletes its oldest photos when less than `-minFreeDiskMB` (256 MB) is free, so the card always holds the most recent photos. Settings, "When the card is full", can stop taking photos instead, so no photo is ever deleted. Status shows how long until the card is full at the current schedule, what happens then, and how many photos were deleted to make room since the camera started.

## Power profiles

Most of the Pi's power goes to hardware that stays on between photos, so the image turns off what the camera never uses. `TIMELAPSEPI_PROFILE` in `src/installer/custompios/config.local` decides about HDMI:

| | `remote` (default) | `bench` |
|---|---|---|
| Use | Unattended deployments | Lab bench, development: a monitor shows the console |
| HDMI | off (`dtoverlay=vc4-kms-v3d,noaudio,nohdmi`) | on (`dtoverlay=vc4-kms-v3d,noaudio`) |

Both profiles:

- `config.txt`: `dtparam=audio=off` (onboard audio), `noaudio` (HDMI audio), `dtoverlay=disable-bt` (Bluetooth; also frees the UART on GPIO 14/15).
- Masked services, if the base image has them: `bluetooth`, `hciuart`, `ModemManager`, `triggerhappy`. `avahi-daemon` stays for the `.local` name.
- The system journal is kept in RAM (`/etc/systemd/journald.conf.d/timelapse.conf`), so it is lost on reboot. The app's own log (`timelapse.log`, `/logs` in the web UI) stays on the SD card, including how the previous run ended (see [Health log](#health-log)). Set `Storage=persistent` there to keep the journal while debugging.
- The CPU runs at its lowest clock: `timelapse-cpufreq.service` sets the `powersave` governor at boot, 600 MHz on a Pi Zero 2 W and 700 MHz on a Pi Zero W instead of up to 1 GHz. Status in the web UI shows the governor and clock. A photo takes a little longer to encode; the Pi idles between photos anyway. The kernel command line (`cpufreq.default_governor=`) cannot do this, because the Raspberry Pi kernel builds the cpufreq driver and every governor as modules. To let the CPU speed up again, run `sudo systemctl disable --now timelapse-cpufreq` and reboot.

HDMI is turned off with the KMS driver's `nohdmi` parameter rather than by dropping the `vc4-kms-v3d` overlay, because that overlay also reserves the memory (CMA) the camera captures into. The older `hdmi_blanking`/`hdmi_ignore_hotplug` settings are for the legacy firmware display driver and have no effect under KMS.

To switch a flashed card, open `config.txt` on the boot partition from any computer and delete or add `,nohdmi` on the `dtoverlay=vc4-kms-v3d` line; the comment above it says which profile the image was built with.

Turning off HDMI saves roughly 20-30 mA, Bluetooth and audio a few mA (estimates; measure with a USB power meter). The WiFi access point remains the largest consumer. For far bigger savings the Pi has to be powered off between photos by a power management board (see [docs/hardware.md](hardware.md)); the image does not support that yet.
