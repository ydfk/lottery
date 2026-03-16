import { defineMock } from "@alova/mock";

export default defineMock({
  // 将路径改回与 API 调用一致
  "/list": ["1", "2", "3", "4", "5"],
  "[POST]/login": { token: "123456", name: "admin", role: "admin" },
});
