<template>
  <div ref="root" class="cselect" :class="{ 'cselect--open': open, 'cselect--disabled': disabled }">
    <button type="button" class="cselect-trigger" :disabled="disabled" @click="toggle">
      <span v-if="selected" class="cselect-value">
        <span v-if="selected.badgeClass" class="cselect-badge" :class="selected.badgeClass">
          {{ selected.label }}
        </span>
        <span v-else>{{ selected.label }}</span>
      </span>
      <span v-else class="cselect-placeholder">{{ placeholder }}</span>
      <i class="pi pi-chevron-down cselect-chevron"></i>
    </button>

    <transition name="cselect-drop">
      <div v-if="open" class="cselect-dropdown">
        <button
          v-for="opt in options"
          :key="opt.value"
          type="button"
          class="cselect-option"
          :class="{ 'cselect-option--active': opt.value === modelValue }"
          @click="pick(opt)"
        >
          <span v-if="opt.badgeClass" class="cselect-badge" :class="opt.badgeClass">{{ opt.label }}</span>
          <span v-else>{{ opt.label }}</span>
          <i v-if="opt.value === modelValue" class="pi pi-check cselect-check"></i>
        </button>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'CustomSelect',
  props: {
    modelValue: { type: String, default: '' },
    options: { type: Array, default: () => [] },
    placeholder: { type: String, default: 'Pilih...' },
    disabled: { type: Boolean, default: false }
  },
  emits: ['update:modelValue'],
  data() {
    return { open: false }
  },
  computed: {
    selected() {
      return this.options.find(o => o.value === this.modelValue) || null
    }
  },
  mounted() {
    document.addEventListener('click', this.onOutside)
  },
  beforeUnmount() {
    document.removeEventListener('click', this.onOutside)
  },
  methods: {
    toggle() {
      this.open = !this.open
    },
    pick(opt) {
      this.$emit('update:modelValue', opt.value)
      this.open = false
    },
    onOutside(e) {
      if (this.$refs.root && !this.$refs.root.contains(e.target)) {
        this.open = false
      }
    }
  }
}
</script>

<style scoped>
.cselect {
  position: relative;
  display: inline-block;
  min-width: 130px;
}

.cselect-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.4rem 0.625rem;
  background: rgba(15, 23, 42, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.8125rem;
  text-align: left;
  transition: border-color 0.15s;
  color: #94a3b8;
}

.cselect-trigger:hover:not(:disabled) {
  border-color: rgba(255, 255, 255, 0.15);
}

.cselect--open .cselect-trigger {
  border-color: rgba(59, 130, 246, 0.4);
}

.cselect--disabled .cselect-trigger {
  opacity: 0.5;
  cursor: not-allowed;
}

.cselect-value {
  flex: 1;
  min-width: 0;
}

.cselect-placeholder {
  color: #64748b;
  flex: 1;
}

.cselect-chevron {
  font-size: 0.625rem;
  color: #64748b;
  transition: transform 0.2s;
  flex-shrink: 0;
}

.cselect--open .cselect-chevron {
  transform: rotate(180deg);
}

/* Badge styling (passed from parent via badgeClass) */
.cselect-badge {
  display: inline-block;
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0.1875rem 0.5rem;
  border-radius: 1rem;
  white-space: nowrap;
}

/* Dropdown */
.cselect-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  min-width: 100%;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.625rem;
  overflow: hidden;
  z-index: 300;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.cselect-option {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.625rem 0.875rem;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 0.8125rem;
  color: #94a3b8;
  text-align: left;
  transition: background 0.12s;
}

.cselect-option:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #f1f5f9;
}

.cselect-option--active {
  background: rgba(59, 130, 246, 0.08);
  color: #f1f5f9;
}

.cselect-check {
  font-size: 0.6875rem;
  color: #3b82f6;
  flex-shrink: 0;
}

/* Transition */
.cselect-drop-enter-active,
.cselect-drop-leave-active {
  transition: opacity 0.12s ease, transform 0.12s ease;
}

.cselect-drop-enter-from,
.cselect-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
