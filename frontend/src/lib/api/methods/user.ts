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

// Logout function (client-side only)
export const logout = (): void => {
  localStorage.removeItem('auth_token');
};