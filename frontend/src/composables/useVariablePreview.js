import { ref, computed } from 'vue'

/**
 * Composable for extracting and rendering {{variable}} placeholders in prompt content.
 * Handles auto-detection, value replacement, and rendered output.
 */
export function useVariablePreview(contentRef) {
  const variableValues = ref({})

  // Extract unique variable names from content
  const variables = computed(() => {
    if (!contentRef.value) return []
    const regex = /\{\{([^}]+)\}\}/g
    const vars = new Set()
    let match
    while ((match = regex.exec(contentRef.value)) !== null) {
      vars.add(match[1].trim())
    }
    return Array.from(vars)
  })

  // Replace all {{var}} with their values in the content
  const renderedContent = computed(() => {
    if (!contentRef.value) return ''
    let result = contentRef.value
    for (const [key, value] of Object.entries(variableValues.value)) {
      if (value) {
        result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value)
      }
    }
    return result
  })

  // Build a highlighted HTML string for preview display
  const highlightedContent = computed(() => {
    if (!contentRef.value) return ''
    return contentRef.value.replace(
      /\{\{([^}]+)\}\}/g,
      (match, varName) => {
        const value = variableValues.value[varName.trim()]
        if (value) {
          return `<mark class="var-filled">${match}</mark>`
        }
        return `<mark class="var-empty">${match}</mark>`
      }
    )
  })

  // Check if a specific variable has a value
  const hasValue = (varName) => {
    return !!variableValues.value[varName]
  }

  // Check if all variables have been filled
  const allFilled = computed(() => {
    return variables.value.length > 0 &&
      variables.value.every(v => !!variableValues.value[v])
  })

  // Fill rate (0-100)
  const fillRate = computed(() => {
    if (variables.value.length === 0) return 100
    const filled = variables.value.filter(v => !!variableValues.value[v]).length
    return Math.round((filled / variables.value.length) * 100)
  })

  // Clear all variable values
  const clearValues = () => {
    variableValues.value = {}
  }

  return {
    variables,
    variableValues,
    renderedContent,
    highlightedContent,
    hasValue,
    allFilled,
    fillRate,
    clearValues
  }
}
