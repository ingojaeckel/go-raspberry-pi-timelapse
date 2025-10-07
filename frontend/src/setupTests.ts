import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

// Mock axios globally
jest.mock('axios');

// Polyfills for Node.js environment
if (typeof global.TextEncoder === 'undefined') {
  const { TextEncoder, TextDecoder } = require('util');
  global.TextEncoder = TextEncoder;
  global.TextDecoder = TextDecoder;
}
