import { computed, reactive, ref } from 'vue'

export function useStandardForm<T extends object>(factory: () => T) {
  const values = reactive(factory()) as T
  const savedValues = ref(JSON.stringify(values))
  const saving = ref(false)
  const error = ref('')
  const dirty = computed(() => JSON.stringify(values) !== savedValues.value)

  function reset(next: T = factory()) {
    Object.assign(values, next)
    savedValues.value = JSON.stringify(values)
    error.value = ''
  }

  async function submit(action: (values: T) => Promise<void>) {
    if (saving.value) return false
    saving.value = true
    error.value = ''
    try {
      await action(values)
      savedValues.value = JSON.stringify(values)
      return true
    } catch (cause) {
      error.value = cause instanceof Error
        ? cause.message
        : 'No se pudo guardar la información.'
      return false
    } finally {
      saving.value = false
    }
  }

  return { values, dirty, saving, error, reset, submit }
}
