# About

This folder manages how disk images are created which can be installed on SD cards to boot Raspberry Pis. For more information checkout https://github.com/guysoft/CustomPiOS.

1. Follow steps in the [CustomPiOS docs](https://github.com/guysoft/CustomPiOS?tab=readme-ov-file#how-to-use-it). This will download the latest OS .img file and prepare it along with other build related files. Optionally: Download another image variant via:

```bash
cd src/image/ && wget https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2021-05-28/2021-05-07-raspios-buster-armhf-lite.zip && cd ../../
```

2. install required software: `sudo apt-get install qemu-user-static p7zip-full`
3. start build: `cd src/ && sudo ./build_dist`. This will create a disk image in `src/workspace/*.img`.
4. The created image can be installed onto an SD card via the [Raspberry Pi Imager](https://www.raspberrypi.org/software/).
