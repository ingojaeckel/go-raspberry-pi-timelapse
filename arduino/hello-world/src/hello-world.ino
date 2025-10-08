/**
 * Hello World Blink Example
 * 
 * This is a simple blink program for the Arduino Giga R1 WiFi board.
 * It blinks the built-in LED on and off every second.
 * 
 * Board: Arduino Giga R1 WiFi
 * LED: Built-in LED (defined by LED_BUILTIN constant)
 * 
 * This sketch is compatible with:
 * - Arduino Giga R1 WiFi (primary target)
 * - Most other Arduino boards (uses standard LED_BUILTIN)
 */

void setup() {
  // Initialize serial communication at 115200 baud rate
  Serial.begin(115200);
  
  // Wait for serial port to connect (useful for debugging)
  while (!Serial && millis() < 3000) {
    ; // wait up to 3 seconds for serial port to connect
  }
  
  // Print welcome message
  Serial.println("Arduino Giga R1 WiFi - Hello World Blink");
  Serial.println("Built-in LED will blink every second");
  
  // Initialize the LED pin as an output
  pinMode(LED_BUILTIN, OUTPUT);
  
  Serial.println("Setup complete!");
}

void loop() {
  // Turn the LED on
  digitalWrite(LED_BUILTIN, HIGH);
  Serial.println("LED ON");
  
  // Wait for 1 second
  delay(1000);
  
  // Turn the LED off
  digitalWrite(LED_BUILTIN, LOW);
  Serial.println("LED OFF");
  
  // Wait for 1 second
  delay(1000);
}
