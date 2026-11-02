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
- The service has no `WorkingDirectory`, so `timelapse.log` and `timelapse-settings.db` are written relative to `/`.

## On-device helper scripts

[`src/installer/device/`](../src/installer/device/) contains scripts to run on an already running Pi:

- `setup.sh`: installs base packages and clones the repo.
- `rtc.sh`: guided setup of the Adafruit PCF8523 real-time clock.
