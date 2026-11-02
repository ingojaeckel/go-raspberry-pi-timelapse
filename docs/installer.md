# Installer: Pi disk image

The [`src/installer/custompios/`](../src/installer/custompios/) folder manages how disk images are created which can be installed on SD cards to boot Raspberry Pis. For more information checkout https://github.com/guysoft/CustomPiOS.

1. Follow steps in the [CustomPiOS docs](https://github.com/guysoft/CustomPiOS?tab=readme-ov-file#how-to-use-it). This will download the latest OS .img file and prepare it along with other build related files. Optionally: Download another image variant via:

```bash
cd src/installer/custompios/image/ && wget https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2021-05-28/2021-05-07-raspios-buster-armhf-lite.zip && cd -
```

2. install required software: `sudo apt-get install qemu-user-static p7zip-full jq`
3. start build: `cd src/installer/custompios/ && sudo ./build_dist`. This will create a disk image in `src/installer/custompios/workspace/*.img`.
4. The created image can be installed onto an SD card via the [Raspberry Pi Imager](https://www.raspberrypi.org/software/).

## How the image finds the app (path coupling)

The image does not contain a separately built binary. Instead:

1. `modules/timelapsepi/start_chroot_script` clones this repo's **default branch from GitHub** to `/home/pi/go-raspberry-pi-timelapse` while the image is built. Your local checkout is not used.
2. `modules/timelapsepi/filesystem/root/etc/systemd/system/timelapse.service` starts the **committed** ARM binary from that clone: `/home/pi/go-raspberry-pi-timelapse/src/backend/bin/arm/go-raspberry-pi-timelapse`.

As a result:

- Rebuild and commit the binary with `make pi` before building an image, or the image ships a stale binary.
- If the binary's location in the repo changes, change `ExecStart` in the same commit. A layout change only reaches new images after it is merged to the default branch.
- The service runs as `pi` with `WorkingDirectory=/home/pi`, so `timelapse.log` (rotated at 10 MB to `timelapse.log.1`) and `timelapse-settings.db` live in `/home/pi`. It restarts automatically (`Restart=always`) after crashes and failed starts.
- `/etc/sudoers.d/010_timelapse` lets `pi` run exactly `/sbin/shutdown -h now` and `/sbin/shutdown -r now`, which the web UI uses to shut down and restart the Pi.
- Images built before these changes ran the service as root without a working directory, so their settings and log are in `/`. The unit file comes from the image, not from the git clone, so `git pull` does not change it; copy the new unit to `/etc/systemd/system/` (and move `/timelapse-settings.db` to `/home/pi`, owned by `pi`) to upgrade such a device.

## On-device helper scripts

[`src/installer/device/`](../src/installer/device/) contains scripts to run on an already running Pi:

- `setup.sh`: installs base packages and clones the repo.
- `rtc.sh`: guided setup of the Adafruit PCF8523 real-time clock.
