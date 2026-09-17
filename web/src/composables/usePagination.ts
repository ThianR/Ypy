import { computed, ref, watch, type Ref } from 'vue'

/** Local, complete datasets only. Remote lists must paginate and filter on the server. */
export function usePagination<T>(items: Ref<T[]>, initialSize = 10) {
  const page = ref(1)
  const pageSize = ref(initialSize)
  const pageCount = computed(() => Math.max(1, Math.ceil(items.value.length / pageSize.value)))
  watch([pageCount, page], () => { page.value = Math.max(1, Math.min(page.value, pageCount.value)) }, { flush: 'sync' })
  watch(pageSize, () => { page.value = 1 }, { flush: 'sync' })
  const start = computed(() => items.value.length ? (page.value - 1) * pageSize.value + 1 : 0)
  const end = computed(() => Math.min(page.value * pageSize.value, items.value.length))
  const rows = computed(() => items.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))
  return { page, pageSize, pageCount, start, end, rows }
}
