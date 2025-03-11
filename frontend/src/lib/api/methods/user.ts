import { alovaInstance } from "..";

// 获取用户信息
export const login = alovaInstance.Post("/login", {
  meta: {
    authRole: "login",
  },
});
