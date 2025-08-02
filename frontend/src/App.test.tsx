import React from 'react';
import { render } from '@testing-library/react';
import App from './App';

test('renders home tab', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/home/i);
  expect(linkElement).toBeInTheDocument();
});

test('renders monitoring tab', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/monitoring/i);
  expect(linkElement).toBeInTheDocument();
});

test('renders settings link', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/settings/i);
  expect(linkElement).toBeInTheDocument();
});