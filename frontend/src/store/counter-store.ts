/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-03-11 11:05:28
 * @LastEditors: ydfk
 * @LastEditTime: 2025-03-11 11:12:19
 */
import { create } from "zustand";
import { immer } from "zustand/middleware/immer";

interface CounterState {
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;
  incrementBy: (value: number) => void;
}

export const useCounterStore = create<CounterState>()(
  immer((set) => ({
    count: 0,
    increment: () =>
      set((state) => {
        state.count += 1;
      }),
    decrement: () =>
      set((state) => {
        state.count -= 1;
      }),
    reset: () =>
      set((state) => {
        state.count = 0;
      }),
    incrementBy: (value) =>
      set((state) => {
        state.count += value;
      }),
  }))
);
