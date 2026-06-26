<template>
  <input
    type="text"
    inputmode="numeric"
    :value="displayValue"
    :placeholder="placeholder"
    @input="handleInput"
    @keydown="handleKeydown"
  />
</template>

<script>
export default {
  name: 'PriceInput',
  props: {
    modelValue: { type: [Number, String], default: '' },
    placeholder: { type: String, default: '0' }
  },
  emits: ['update:modelValue'],
  computed: {
    displayValue() {
      if (this.modelValue === '' || this.modelValue === null || this.modelValue === undefined) return ''
      const n = Number(this.modelValue)
      if (isNaN(n) || n <= 0) return ''
      return n.toLocaleString('id-ID')
    }
  },
  methods: {
    handleInput(e) {
      const raw = e.target.value.replace(/[^\d]/g, '')
      if (!raw) {
        this.$emit('update:modelValue', '')
        e.target.value = ''
        return
      }
      const n = parseInt(raw, 10)
      this.$emit('update:modelValue', n)
      const formatted = n.toLocaleString('id-ID')
      const cursorFromEnd = e.target.value.length - e.target.selectionEnd
      e.target.value = formatted
      const newPos = Math.max(0, formatted.length - cursorFromEnd)
      e.target.setSelectionRange(newPos, newPos)
    },
    handleKeydown(e) {
      if (
        e.key === 'Backspace' || e.key === 'Delete' ||
        e.key === 'Tab' || e.key === 'Escape' || e.key === 'Enter' ||
        e.key === 'Home' || e.key === 'End' ||
        e.key.startsWith('Arrow') ||
        e.ctrlKey || e.metaKey
      ) return
      if (!/^\d$/.test(e.key)) e.preventDefault()
    }
  }
}
</script>
