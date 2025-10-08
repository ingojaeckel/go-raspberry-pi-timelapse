# Hello World - Arduino Giga R1 WiFi

A simple blink program for the [Arduino Giga R1 WiFi](https://store.arduino.cc/products/giga-r1-wifi) board.

## Board Specifications

- **Microcontroller**: STM32H747XI dual-core
  - Cortex-M7 core at 480 MHz
  - Cortex-M4 core at 240 MHz
- **WiFi/Bluetooth**: Murata 1DX module
- **USB**: USB-C connector
- **Built-in LED**: Pin LED_BUILTIN (typically pin 13 on Arduino boards)

## What It Does

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

## Configuration

The project uses `arduino-cli.yaml` for build configuration:

```yaml
# Arduino CLI Project Configuration
fqbn: arduino:mbed_giga:giga
```

This specifies the target board as Arduino Giga R1 WiFi.

## Building and Deploying

### Using Arduino CLI (Recommended)

1. **Build the project:**
   ```bash
   cd arduino/hello-world
   ./scripts/build.sh
   ```
   
   The script will automatically:
   - Download and install Arduino CLI if not present
   - Read configuration from `arduino-cli.yaml`
   - Install the Arduino Mbed OS Giga Boards package
   - Compile the sketch for the Giga R1 WiFi

2. **Deploy to board:**
   ```bash
   ./scripts/deploy.sh
   ```
   
   Or specify a port manually:
   ```bash
   ./scripts/deploy.sh /dev/ttyACM0
   ```

3. **Monitor serial output:**
   ```bash
   ./scripts/monitor.sh
   ```

### Using Arduino IDE

1. **Open Arduino IDE**

2. **Install board support:**
   - Go to Tools → Board → Boards Manager
   - Search for "Arduino Giga" or "Mbed Giga"
   - Install "Arduino Mbed OS Giga Boards"

3. **Open the sketch:**
   - File → Open
   - Navigate to `arduino/hello-world/src/hello-world.ino`

4. **Select board:**
   - Tools → Board → Arduino Mbed OS Giga Boards → Arduino Giga R1 WiFi

5. **Select port:**
   - Tools → Port → Select your board's port

6. **Upload:**
   - Click the Upload button or Sketch → Upload

## Troubleshooting

### Arduino CLI Issues

- **Download failed:** Check internet connection or download manually from https://arduino.github.io/arduino-cli/
- **Board package not found:** Run `arduino-cli core update-index`
- **Upload failed:** 
  - Check USB connection
  - Try `arduino-cli board list` to see available boards
  - Specify port manually: `./scripts/deploy.sh /dev/ttyACM0`

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

## Compatibility

This sketch is compatible with:
- ✅ Arduino Giga R1 WiFi (primary target)
- ✅ Most Arduino boards (uses standard LED_BUILTIN)
- ✅ Arduino IDE 1.8.x and 2.x
- ✅ Arduino CLI

## Resources

- [Arduino Giga R1 WiFi Official Page](https://store.arduino.cc/products/giga-r1-wifi)
- [Arduino Giga R1 WiFi Documentation](https://docs.arduino.cc/hardware/giga-r1-wifi)
- [Arduino CLI Documentation](https://arduino.github.io/arduino-cli/)
- [Arduino Language Reference](https://www.arduino.cc/reference/en/)
- [STM32H747 Datasheet](https://www.st.com/en/microcontrollers-microprocessors/stm32h747xi.html)
