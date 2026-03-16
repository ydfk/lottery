import { createAlova } from "alova";
import { createServerTokenAuthentication } from "alova/client";
import ReactHook from "alova/react";
import adapterFetch from "alova/fetch";
import { mockAdapter } from "./mock/index";

const { onAuthRequired, onResponseRefreshToken } = createServerTokenAuthentication({
  async login(response) {
    const data = await response.clone().json();
    localStorage.setItem("token", data.token);
  },
  assignToken: (method) => {
    method.config.headers.Authorization = localStorage.getItem("token");
  },
  logout() {
    localStorage.removeItem("token");
  },
});

const useMock = import.meta.env.VITE_USE_MOCK === "true";
console.log("🚀 ~ useMock:", useMock);

export const alovaInstance = createAlova({
  baseURL: "/api",
  statesHook: ReactHook,
  requestAdapter: useMock ? mockAdapter : adapterFetch(),
  beforeRequest: onAuthRequired(),
  responded: onResponseRefreshToken(),
});
