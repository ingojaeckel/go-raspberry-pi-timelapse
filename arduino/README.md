# Arduino Projects

This directory contains Arduino projects for various boards.

## Project Structure

Each project is in its own subdirectory with the following structure:

```
arduino/
├── hello-world/                   # Example: Hello World blink project
│   ├── arduino-cli.yaml          # Project configuration (FQBN, libraries, etc.)
│   ├── hello-world/              # Sketch directory (must match sketch name)
│   │   └── hello-world.ino       # Arduino sketch
│   └── scripts/
│       ├── build.sh              # Build script
│       ├── deploy.sh             # Deploy script
│       └── monitor.sh            # Serial monitor script
└── README.md                      # This file
```

## Projects

### hello-world

A simple blink program for the **Arduino Giga R1 WiFi** board that:
- Blinks the built-in LED every second
- Prints status messages to the serial monitor at 115200 baud

**Board:** Arduino Giga R1 WiFi (STM32H747XI dual-core)

See [hello-world/README.md](hello-world/README.md) for details.

## Adding New Projects

To add a new project:

1. Create a new directory: `arduino/my-project/`
2. Create a sketch directory matching your sketch name: `arduino/my-project/my-sketch/`
3. Create your `.ino` file: `arduino/my-project/my-sketch/my-sketch.ino`
4. Create an `arduino-cli.yaml` configuration file:
   ```yaml
   # Arduino CLI Project Configuration
   fqbn: arduino:mbed_giga:giga  # Your target board FQBN
   ```
5. Copy and adapt the build scripts from an existing project
6. Add your project to the GitHub Actions workflow matrix in `.github/workflows/arduino.yml`

**Note:** Arduino CLI requires the sketch directory name to match the .ino filename (e.g., `my-sketch/my-sketch.ino`).

The configuration file supports:
- **fqbn**: Fully Qualified Board Name (required)
- **libraries**: List of library dependencies (optional)
- **build_properties**: Additional build properties (optional)

## Prerequisites

### Arduino CLI (Recommended)

- **Arduino CLI** (auto-installed by build scripts)
- Works with any Arduino-compatible board
- Best for local development and deployment

### Arduino IDE

- Arduino IDE 2.0 or later
- Install required board packages from Boards Manager
- Easy to use GUI interface

## Building and Deploying

Each project has its own build scripts:

```bash
cd arduino/hello-world
./scripts/build.sh      # Build the project
./scripts/deploy.sh     # Deploy to connected board
./scripts/monitor.sh    # Monitor serial output
```

The build script automatically:
- Installs Arduino CLI if not present
- Reads the project configuration from `arduino-cli.yaml`
- Installs the required board platform
- Compiles the sketch for the target board

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/arduino.yml`):
- Automatically builds all projects on code changes
- Reads each project's `arduino-cli.yaml` for board configuration
- Creates firmware artifacts for each project
- Uses a matrix strategy to build multiple projects in parallel

To add a project to CI/CD, add it to the workflow matrix:

```yaml
strategy:
  matrix:
    project:
      - hello-world
      - my-new-project  # Add your project here
```

## Resources

- [Arduino CLI Documentation](https://arduino.github.io/arduino-cli/)
- [Arduino Language Reference](https://www.arduino.cc/reference/en/)
