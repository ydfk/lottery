import { useRequest, useFetcher, usePagination } from "alova/client";
import { useCallback } from "react";
import { getLotteryTypes, getRecommendations } from "../lib/api/methods/lottery";

export function useLotteryData() {
  const {
    // 加载状态
    loading: isLoadingRecommendations,
    // 列表数据
    data: recommendations = [],
    error: recommendationsError,
    // 是否为最后一页
    // 下拉加载时可通过此参数判断是否还需要加载
    isLastPage,
    // 当前页码，改变此页码将自动触发请求
    page,
    // 每页数据条数
    pageSize,
    // 分页页数
    pageCount,
    // 总数据量
    total,
    // 更新状态
    update,
  } = usePagination((page, pageSize) => getRecommendations(page, pageSize), {
    append: true, // 是否追加数据
    // 请求前的初始数据（接口返回的数据格式）
    initialData: {
      total: 0,
      data: [],
      hasMore: false,
    },
    initialPage: 1, // 初始页码，默认为1
    initialPageSize: 20, // 初始每页数据条数，默认为10
  });

  // 使用 useRequest 获取彩票类型
  const {
    loading: isLoadingTypes,
    data: lotteryTypes = [],
    error: typesError,
  } = useRequest(getLotteryTypes, { immediate: true });

  // 更新购买状态
  const updatePurchaseStatus = async (id: number, isPurchased: boolean) => {
    try {
      await updatePurchaseStatus(id, isPurchased)

      const newData = [...recommendations];
        
      // 查找并更新对应ID的彩票数据
      const index = newData.findIndex(item => item.id === id);
      if (index !== -1) {
        newData[index] = {
          ...newData[index],
          isPurchased: isPurchased
        };
      }

      //更新列表中数据状态
      update({
        data: newData
      });

      return true;
    } catch (error) {
      console.error("更新购买状态失败", error);
      return false;
    }
  };

  // 加载更多数据
  const loadMore = useCallback(() => {
    if (!isLoadingRecommendations && !isLastPage) {
      update({
        page: page + 1,
      });
    }
  }, [isLoadingRecommendations, isLastPage, page]);

  // 重置并重新加载数据
  const refresh = useCallback(() => {
    update({
      page: 1,
    });
  }, [page]);

  // 获取彩票类型名称的工具函数
  const getLotteryTypeName = useCallback(
    (typeId: number) => {
      const type = lotteryTypes.find((type) => type.id === typeId);
      return type ? type.name : `类型 ${typeId}`;
    },
    [lotteryTypes]
  );

  return {
    recommendations,
    isLoading: isLoadingRecommendations,
    error: recommendationsError,
    lotteryTypes,
    isLoadingTypes,
    typesError,
    hasMore: !isLastPage || false,
    total: total || 0,
    loadMore,
    refresh,
    updatePurchaseStatus,
    getLotteryTypeName,
  };
}
