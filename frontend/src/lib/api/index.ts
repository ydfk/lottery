import { createAlova } from 'alova';
import ReactHook from 'alova/react';
import adapterFetch from 'alova/fetch';

// Create alova instance
export const alovaInstance = createAlova({
  baseURL: "/api",
  statesHook: ReactHook,
  requestAdapter: adapterFetch(),
  beforeRequest: ({config}) => {
    // Get token from localStorage
    const token = localStorage.getItem('auth_token');
    
    // If token exists, add it to the authorization header
    if (token) {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      };
    }
  },
  responded: {
    // Handle common response
    onSuccess: async (response) => {
      if (!response.ok) {
        // If response status is not ok, throw error
        const error = await response.json();
        return Promise.reject(error);
      }
      return response.json();
    },
    onError: (error) => {
      console.error('Request error:', error);
      return Promise.reject(error);
    },
},
});

export default alovaInstance;