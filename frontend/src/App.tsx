/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-03-11 10:49:14
 * @LastEditors: ydfk
 * @LastEditTime: 2025-03-11 11:11:01
 */
import { useState } from "react";
import { ThemeProvider } from "./components/theme-provider";
import { ThemeToggle } from "./components/theme-toggle";
import { ZustandCounter } from "./components/zustand-counter";
import { Button } from "./components/ui/button";
import { useRequest } from 'alova/client';
import { login } from "./lib/api/methods/user";

function App() {
  const [count, setCount] = useState(0);
  const { loading, data, send } = useRequest(login, {
    immediate: true
  });

  return (
    <ThemeProvider>
      <div className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white flex flex-col items-center justify-center p-4">
        <div className="absolute top-4 right-4">
          <ThemeToggle />
        </div>

        <div className="w-full max-w-md space-y-8 text-center">
          <h1 className="text-4xl font-bold">React 19 + TypeScript</h1>
          <p className="text-lg text-gray-500 dark:text-gray-400">使用最新技术栈构建的现代化React应用</p>

          <div className="flex flex-col items-center space-y-4 p-6 border rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white shadow-sm">
            <h2 className="text-2xl font-semibold">计数器示例</h2>
            <p className="text-3xl font-bold">{count}</p>
            <div className="flex space-x-2">
              <Button onClick={() => setCount(count - 1)}>减少</Button>
              <Button variant="outline" onClick={() => setCount(0)}>
                重置
              </Button>
              <Button onClick={() => setCount(count + 1)}>增加</Button>
            </div>
          </div>

          <ZustandCounter />

          <div className="flex flex-col items-center space-y-4 p-6 border rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white shadow-sm">
            <h2 className="text-2xl font-semibold">Alova示例</h2>
            {loading ? (
              <p>加载中...</p>
            ) : (
              <div className="text-left w-full">
                <pre className="bg-gray-100 dark:bg-gray-700 p-4 rounded-md overflow-auto">
                  {JSON.stringify(data, null, 2)}
                </pre>
              </div>
            )}
            <Button onClick={() => send()} disabled={loading}>
              {loading ? "加载中..." : "刷新数据"}
            </Button>
          </div>

          <p className="text-sm text-gray-500 dark:text-gray-400">
            技术栈: React 19, TypeScript, Vite 6, Alova, Zustand, shadcn/ui
          </p>
        </div>
      </div>
    </ThemeProvider>
  );
}

export default App;
