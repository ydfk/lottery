import { createAlova } from 'alova';
import ReactHook from 'alova/react';
import adapterFetch from 'alova/fetch';
import { createServerTokenAuthentication } from 'alova/client';

const { onAuthRequired, onResponseRefreshToken } = createServerTokenAuthentication({
  async login(response, method) {
    const token = (await response.clone().json()).token;
    localStorage.setItem('auth_token', token);
  },
  assignToken: method => {
    method.config.headers.Authorization = `Bearer ${localStorage.getItem('auth_token')}`;
  },
});

// 创建 alova 实例
export const alovaInstance = createAlova({
  baseURL: "/api",
  statesHook: ReactHook,
  requestAdapter: adapterFetch(),
  beforeRequest: onAuthRequired(({config}) => {
  }),
  responded: onResponseRefreshToken ({
    // 处理通用响应
    onSuccess: async (response) => {
      if (!response.ok) {
        // 如果响应状态不是 ok，抛出错误
        const error = await response.json();
        return Promise.reject(error);
      }
      return response.json();
    },
    onError: (error) => {
      console.error('请求错误:', error);
      return Promise.reject(error);
    },
  }),
});

export default alovaInstance;