// For production builds, the frontend runs embedded into the backend.
// In that case requests can be send to the server hosting the frontend. No base URL is required.
// In non-production builds, the frontend can retrieve data from a separately running backend service running on localhost:8080.
export const BaseUrl = process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8080'