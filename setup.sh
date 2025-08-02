#!/usr/bin/env bash
#
# You might want to download & execute this to start the installation via:
# wget -O - https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/setup.sh | bash
#
set -e -x

# Try to free up some space now before installing/updating may need it
sudo apt-get autoclean
sudo apt-get autoremove
sudo apt-get clean

sudo apt-get update -y
sudo apt-get upgrade -y

# Install basic dependencies
sudo apt-get install -y htop screen git python3-smbus i2c-tools

# Install OpenCV C++ libraries and development headers for GoCV
sudo apt-get install -y libopencv-dev libopencv-contrib-dev
sudo apt-get install -y build-essential cmake pkg-config
sudo apt-get install -y libjpeg-dev libtiff5-dev libpng-dev
sudo apt-get install -y libavcodec-dev libavformat-dev libswscale-dev
sudo apt-get install -y libgtk2.0-dev libcanberra-gtk-module

# Install Python OpenCV for fallback compatibility (optional)
sudo apt-get install -y python3-pip python3-opencv python3-numpy
pip3 install --user opencv-python numpy requests

git clone https://github.com/ingojaeckel/go-raspberry-pi-timelapse.git
cd go-raspberry-pi-timelapse/

# Make detection scripts executable
chmod +x detection/download_models.sh

# Download YOLO models for offline object detection
echo "Setting up advanced object detection models..."
./detection/download_models.sh

echo "Setup complete! The system now uses native Go OpenCV bindings for object detection."
echo "Python OpenCV support is available as fallback, but primary detection uses GoCV."
