# About

This folder manages how disk images are created which can be installed on SD cards to boot Raspberry Pis. For more information checkout https://github.com/guysoft/CustomPiOS.

To create a new disk image, run the following from Linux (after going through the CustomPiOS setup steps). This will create a disk image in `src/workspace/*.img`. The created image can be installed onto an SD card via the [Raspberry Pi Imager](https://www.raspberrypi.org/software/).

```
sudo src/build_dist raspios_lite_arm64
```
