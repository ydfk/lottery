import { createAlova } from 'alova';
import ReactHook from 'alova/react';
import adapterFetch from 'alova/fetch';

// 创建 alova 实例
export const alovaInstance = createAlova({
  baseURL: "/api",
  statesHook: ReactHook,
  requestAdapter: adapterFetch(),
  beforeRequest: ({config}) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('auth_token');
    
    // 如果 token 存在，添加到请求头
    if (token) {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      };
    }
  },
  responded: {
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
  },
});

export default alovaInstance;