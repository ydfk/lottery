import { calculateEntriesCost } from "./lottery-cost";

test("双色球和大乐透普通投注按每注 2 元计算", () => {
  expect(calculateEntriesCost([{ multiple: 3, isAdditional: false }])).toBe(6);
});

test("大乐透追加和多注倍数正确合计", () => {
  expect(
    calculateEntriesCost([
      { multiple: 2, isAdditional: false },
      { multiple: 3, isAdditional: true },
    ])
  ).toBe(13);
});

test("倍数按 1 到 99 的边界计算", () => {
  expect(calculateEntriesCost([{ multiple: 0, isAdditional: false }])).toBe(2);
  expect(calculateEntriesCost([{ multiple: 100, isAdditional: false }])).toBe(198);
});
