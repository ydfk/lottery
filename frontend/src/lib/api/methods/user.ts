import {alovaInstance} from '../index';

interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  token: string;
}

// Login API method
export const login = (data: LoginRequest) => {
  return alovaInstance.Post<LoginResponse>('/login', data);
};

// Check if user is authenticated
export const isAuthenticated = (): boolean => {
  return !!localStorage.getItem('auth_token');
};

// Logout function (client-side only)
export const logout = (): void => {
  localStorage.removeItem('auth_token');
};