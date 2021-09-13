# About

This folder manages how disk images are created which can be installed on SD cards to boot Raspberry Pis. For more information checkout https://github.com/guysoft/CustomPiOS.

To create a new disk image, run the following from Linux (after going through the CustomPiOS setup steps - specifically, you may need to install the following: `sudo apt-get install qemu-user-static`). This will create a disk image in `src/workspace/*.img`. The created image can be installed onto an SD card via the [Raspberry Pi Imager](https://www.raspberrypi.org/software/).

```
cd src/ && sudo ./build_dist
```

Optionally: Download another image variant before invoking `build_dist`
```
cd src/image/ && wget https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2021-05-28/2021-05-07-raspios-buster-armhf-lite.zip && cd ../../
```
