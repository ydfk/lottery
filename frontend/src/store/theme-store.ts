/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-03-11 11:00:18
 * @LastEditors: ydfk
 * @LastEditTime: 2025-03-11 11:12:05
 */
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

interface ThemeState {
  theme: "light" | "dark";
  toggleTheme: () => void;
  setTheme: (theme: "light" | "dark") => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    immer((set) => ({
      theme: "light",
      toggleTheme: () =>
        set((state) => {
          state.theme = state.theme === "light" ? "dark" : "light";
        }),
      setTheme: (theme) =>
        set((state) => {
          state.theme = theme;
        }),
    })),
    {
      name: "theme-storage",
    }
  )
);
