import { ref, computed, watch } from 'vue'

export function usePagination(fetchFn, options = {}) {
  const {
    pageSize = 12,
    immediate = true
  } = options

  const currentPage = ref(1)
  const total = ref(0)
  const loading = ref(false)
  const data = ref([])

  const totalPages = computed(() => Math.ceil(total.value / pageSize))

  const fetch = async (extraParams = {}) => {
    loading.value = true
    try {
      const params = {
        page: currentPage.value,
        limit: pageSize,
        ...extraParams
      }
      const res = await fetchFn(params)
      if (res.data.success) {
        data.value = res.data.data || []
        if (res.data.meta) {
          total.value = res.data.meta.total
        }
      }
    } finally {
      loading.value = false
    }
  }

  const reset = () => {
    currentPage.value = 1
    data.value = []
    total.value = 0
  }

  watch(currentPage, () => fetch())

  if (immediate) {
    fetch()
  }

  return {
    data,
    currentPage,
    total,
    totalPages,
    pageSize,
    loading,
    fetch,
    reset
  }
}
