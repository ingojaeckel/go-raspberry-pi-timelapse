import React from 'react';
import { render } from '@testing-library/react';
import App from './App';

test('renders Photos link', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/Photos/i);
  expect(linkElement).toBeInTheDocument();
});

test('renders Monitoring link', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/Monitoring/i);
  expect(linkElement).toBeInTheDocument();
});

test('renders Settings link', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/Settings/i);
  expect(linkElement).toBeInTheDocument();
});