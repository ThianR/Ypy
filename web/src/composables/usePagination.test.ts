import { computed, ref } from 'vue'
import { describe, expect, it } from 'vitest'
import { usePagination } from './usePagination'

describe('local pagination', () => {
  it('bounds a thousand records and includes matches outside the first page', () => {
    const all = Array.from({ length: 1000 }, (_, id) => ({ id }))
    const query = ref('')
    const filtered = computed(() => all.filter(row => String(row.id).includes(query.value)))
    const pager = usePagination(filtered)
    expect(pager.rows.value).toHaveLength(10)
    pager.page.value = 100
    expect(pager.rows.value[9].id).toBe(999)
    query.value = '999'
    expect(pager.page.value).toBe(1)
    expect(pager.rows.value).toEqual([{ id: 999 }])
  })
  it('recovers from deleting the last row of the last page and empty results', () => {
    const items = ref(Array.from({ length: 11 }, (_, i) => i))
    const pager = usePagination(items)
    pager.page.value = 2
    items.value.pop()
    expect(pager.page.value).toBe(1)
    expect(pager.rows.value).toHaveLength(10)
    items.value = []
    expect([pager.start.value, pager.end.value, pager.pageCount.value]).toEqual([0, 0, 1])
  })
  it('resets page when changing page size', () => {
    const pager = usePagination(ref(Array.from({ length: 80 }, (_, i) => i)))
    pager.page.value = 3
    pager.pageSize.value = 25
    expect(pager.page.value).toBe(1)
    expect(pager.rows.value).toHaveLength(25)
  })
})
