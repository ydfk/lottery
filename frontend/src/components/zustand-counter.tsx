/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-03-11 11:05:46
 * @LastEditors: ydfk
 * @LastEditTime: 2025-03-11 11:05:56
 */
import { Button } from "./ui/button";
import { useCounterStore } from "../store/counter-store";

export function ZustandCounter() {
  const { count, increment, decrement, reset, incrementBy } = useCounterStore();

  return (
    <div className="flex flex-col items-center space-y-4 p-6 border rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white shadow-sm">
      <h2 className="text-2xl font-semibold">Zustand状态管理示例</h2>
      <p className="text-3xl font-bold">{count}</p>
      <div className="flex space-x-2">
        <Button onClick={decrement}>减少</Button>
        <Button variant="outline" onClick={reset}>
          重置
        </Button>
        <Button onClick={increment}>增加</Button>
      </div>
      <div className="flex space-x-2">
        <Button variant="secondary" onClick={() => incrementBy(5)}>
          +5
        </Button>
        <Button variant="secondary" onClick={() => incrementBy(10)}>
          +10
        </Button>
      </div>
      <p className="text-sm text-gray-500 dark:text-gray-400">这个计数器使用Zustand进行状态管理</p>
    </div>
  );
}
