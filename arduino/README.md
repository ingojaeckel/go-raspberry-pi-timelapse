# Arduino Giga R1 WiFi Projects

This directory contains Arduino projects for the [Arduino Giga R1 WiFi](https://store.arduino.cc/products/giga-r1-wifi) board.

## Board Specifications

- **Microcontroller**: STM32H747XI dual-core
  - Cortex-M7 core at 480 MHz
  - Cortex-M4 core at 240 MHz
- **WiFi/Bluetooth**: Murata 1DX module
- **USB**: USB-C connector
- **Built-in LED**: Pin LED_BUILTIN (typically pin 13 on Arduino boards)

## Project Structure

```
arduino/
├── platformio.ini              # PlatformIO configuration (for CI/CD syntax validation)
├── src/
│   └── hello-world.ino        # Hello World blink sketch
├── scripts/
│   ├── build.sh               # Build script (PlatformIO)
│   ├── build-arduino-cli.sh   # Build script (Arduino CLI - RECOMMENDED)
│   ├── deploy.sh              # Upload/deploy script (PlatformIO)
│   ├── deploy-arduino-cli.sh  # Upload/deploy script (Arduino CLI - RECOMMENDED)
│   ├── monitor.sh             # Serial monitor script (PlatformIO)
│   └── monitor-arduino-cli.sh # Serial monitor script (Arduino CLI - RECOMMENDED)
└── README.md                  # This file
```

This structure supports adding multiple projects in the future while using the same Giga R1 WiFi platform configuration.

## Prerequisites

### Option 1: Arduino CLI (Recommended for Development)

- **Arduino CLI** (auto-installed by build-arduino-cli.sh)
- Works with Arduino Giga R1 WiFi board package
- Best for local development and deployment

### Option 2: Arduino IDE (Best for Beginners)

- Arduino IDE 2.0 or later
- Arduino Mbed OS Giga Boards package
- Easy to use GUI interface

### Option 3: PlatformIO (For CI/CD)

- Python 3.6 or later
- PlatformIO CLI
- Used for automated builds in GitHub Actions

## Getting Started

### Using Arduino CLI (Recommended)

Arduino CLI provides the best support for the Arduino Giga R1 WiFi board.

1. **Build the project:**
   ```bash
   cd arduino
   ./scripts/build-arduino-cli.sh
   ```
   
   The script will automatically:
   - Download and install Arduino CLI if not present
   - Install the Arduino Mbed OS Giga Boards package
   - Compile the sketch for the Giga R1 WiFi

2. **Deploy to board:**
   ```bash
   ./scripts/deploy-arduino-cli.sh
   ```
   
   Or specify a port manually:
   ```bash
   ./scripts/deploy-arduino-cli.sh /dev/ttyACM0
   ```

3. **Monitor serial output:**
   ```bash
   ./scripts/monitor-arduino-cli.sh
   ```

### Using Arduino IDE

1. **Open Arduino IDE**

2. **Install board support:**
   - Go to Tools → Board → Boards Manager
   - Search for "Arduino Giga" or "Mbed Giga"
   - Install "Arduino Mbed OS Giga Boards"

3. **Open the sketch:**
   - File → Open
   - Navigate to `arduino/src/hello-world.ino`

4. **Select board:**
   - Tools → Board → Arduino Mbed OS Giga Boards → Arduino Giga R1 WiFi

5. **Select port:**
   - Tools → Port → Select your board's port

6. **Upload:**
   - Click the Upload button or Sketch → Upload

The sketch will compile and upload to your board.

### Using PlatformIO (CI/CD)

PlatformIO is configured primarily for CI/CD pipelines. For local development, use Arduino CLI or Arduino IDE.

```bash
cd arduino
./scripts/build.sh
```

**Note:** The PlatformIO configuration uses a compatible board for build validation. For actual Giga R1 WiFi deployment, use Arduino CLI or Arduino IDE.

## Hello World Blink Project

The `hello-world.ino` sketch is a simple blink program that:
- Blinks the built-in LED on and off every second
- Prints status messages to the serial monitor at 115200 baud
- Demonstrates basic Arduino programming for the Giga R1 WiFi

### Expected Output

When running, you should see the built-in LED blinking and serial output:
```
Arduino Giga R1 WiFi - Hello World Blink
Built-in LED will blink every second
Setup complete!
LED ON
LED OFF
LED ON
LED OFF
...
```

## Build & Deploy Options Summary

| Method | Best For | Build Command | Deploy Command |
|--------|----------|---------------|----------------|
| **Arduino CLI** | Development & Production | `./scripts/build-arduino-cli.sh` | `./scripts/deploy-arduino-cli.sh` |
| **Arduino IDE** | Beginners & Quick Testing | GUI Upload Button | GUI Upload Button |
| **PlatformIO** | CI/CD & Automation | `./scripts/build.sh` | `./scripts/deploy.sh` |

## CI/CD Integration

The project includes a GitHub Actions workflow (`.github/workflows/arduino.yml`) that:
- Automatically builds the project on code changes
- Validates the Arduino sketch syntax
- Creates firmware artifacts
- Works without physical hardware

## Adding New Projects

To add a new project for the Giga R1 WiFi:

1. Create a new directory: `arduino/my-project/`
2. Create a `src/` directory with your `.ino` sketch
3. Copy and adapt the build scripts from the hello-world project
4. For Arduino CLI: Use the same `build-arduino-cli.sh` pattern
5. For Arduino IDE: Open the sketch directly

The modular structure allows multiple independent projects while sharing the same board platform.

## Troubleshooting

### Arduino CLI Issues

- **Download failed:** Check internet connection or download manually from https://arduino.github.io/arduino-cli/
- **Board package not found:** Run `arduino-cli core update-index`
- **Upload failed:** 
  - Check USB connection
  - Try `arduino-cli board list` to see available boards
  - Specify port manually: `./scripts/deploy-arduino-cli.sh /dev/ttyACM0`

### PlatformIO Issues

- **Platform not found:** Ensure internet connection for package downloads
- **Board not found:** PlatformIO config uses a compatible board for validation
- **For actual deployment:** Use Arduino CLI or Arduino IDE

### Arduino IDE Issues

- **Board not showing:** 
  - Ensure "Arduino Mbed OS Giga Boards" package is installed
  - Restart Arduino IDE
- **Upload failed:** 
  - Try pressing reset button twice quickly to enter bootloader
  - Check cable connection
  - Verify correct port is selected

### Permission Issues (Linux)

```bash
# Add user to dialout group for serial port access
sudo usermod -a -G dialout $USER
# Log out and back in for changes to take effect
```

## Sketch Compatibility

The hello-world.ino sketch is designed to be compatible with:
- ✅ Arduino Giga R1 WiFi (primary target)
- ✅ Most Arduino boards (uses standard LED_BUILTIN)
- ✅ Arduino IDE 1.8.x and 2.x
- ✅ Arduino CLI
- ✅ PlatformIO (with compatible board selection)

## Resources

- [Arduino Giga R1 WiFi Official Page](https://store.arduino.cc/products/giga-r1-wifi)
- [Arduino Giga R1 WiFi Documentation](https://docs.arduino.cc/hardware/giga-r1-wifi)
- [Arduino CLI Documentation](https://arduino.github.io/arduino-cli/)
- [PlatformIO Documentation](https://docs.platformio.org/)
- [Arduino Language Reference](https://www.arduino.cc/reference/en/)
- [STM32H747 Datasheet](https://www.st.com/en/microcontrollers-microprocessors/stm32h747xi.html)
