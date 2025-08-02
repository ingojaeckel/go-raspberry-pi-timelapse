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

# Install OpenCV and Python dependencies for object detection
sudo apt-get install -y python3-pip python3-opencv python3-numpy
sudo apt-get install -y libopencv-dev python3-dev

# Install additional Python packages for enhanced object detection
pip3 install --user opencv-python numpy requests

git clone https://github.com/ingojaeckel/go-raspberry-pi-timelapse.git
cd go-raspberry-pi-timelapse/

# Make detection scripts executable
chmod +x detection/download_models.sh
chmod +x detection/opencv_detector.py

# Download YOLO models for offline object detection
echo "Setting up advanced object detection models..."
./detection/download_models.sh
