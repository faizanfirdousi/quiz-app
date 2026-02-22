import axios from 'axios';
import { getIdToken, refreshSession } from '../auth/cognitoClient';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: attach Bearer token
apiClient.interceptors.request.use(
  async (config) => {
    try {
      const token = await getIdToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    } catch (err) {
      console.warn('Failed to get auth token:', err);
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: handle auth errors
apiClient.interceptors.response.use(
  (response) => {
    const requestId = response.data?.requestId;
    if (requestId) {
      console.debug(`[API] requestId: ${requestId}`);
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // On 401: attempt token refresh, retry once
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        const newSession = await refreshSession();
        if (newSession) {
          const newToken = newSession.getIdToken().getJwtToken();
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError);
      }
      // Redirect to login
      window.location.href = '/login';
      return Promise.reject(error);
    }

    // On 403: redirect home
    if (error.response?.status === 403) {
      console.error('Access denied');
      window.location.href = '/';
    }

    return Promise.reject(error);
  }
);

export default apiClient;
