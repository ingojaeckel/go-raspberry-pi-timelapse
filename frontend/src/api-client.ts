/**
 * API Client for the Go Raspberry Pi Timelapse API
 * Auto-generated types from OpenAPI specification
 */
import createClient from 'openapi-fetch';
import type { paths } from './generated/api';
import { BaseUrl } from './conf/config';

// Create the API client with the base URL
export const apiClient = createClient<paths>({ baseUrl: BaseUrl });

// Export types for convenience
export type { paths, components } from './generated/api';
