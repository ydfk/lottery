import adapterFetch from "alova/fetch";
import { createAlovaMockAdapter } from "@alova/mock";
import userMock from "./user";

export const mockAdapter = createAlovaMockAdapter([userMock], {
  httpAdapter: adapterFetch(),
  mockRequestLogger: true,
  delay: 1000,
  matchMode: "methodurl",
});
