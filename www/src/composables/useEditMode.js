/**
 * Composable: Data Tab Inline Editing
 * Handles edit mode, cell editing, row selection, new rows, delete, and batch save.
 *
 * Usage:
 *   const edit = useEditMode(t, { ensureDataTabSchema, refreshTableData })
 */
import { nextTick } from 'vue'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

/**
 * @param {Function} t - i18n translate function
 * @param {Object} deps - late-bound dependencies (set via deps.xxx = fn after init)
 * @param {Function} deps.ensureDataTabSchema
 * @param {Function} deps.refreshTableData
 */
export function useEditMode(t, deps = {}) {
  // ── Cell Editing ──
  async function startEdit(tab, rowIdx, col) {
    if (tab.editingCell?.rowIdx === rowIdx && tab.editingCell?.col === col) return

    if (tab.canEdit == null) {
      try {
        await deps.ensureDataTabSchema(tab)
      } catch (error) {
        gmssh.error(error.message)
        return
      }
    }

    if (!tab.canEdit) {
      gmssh.error(t('dataTab.noPKError'))
      return
    }

    tab.editingCell = { rowIdx, col }
  }

  function commitEdit(tab, rowIdx, col, newValue) {
    const originalValue = tab.tableData?.rows?.[rowIdx]?.[col]

    if (String(newValue) !== String(originalValue ?? '')) {
      if (!tab.pendingEdits[rowIdx]) {
        tab.pendingEdits[rowIdx] = {}
      }
      tab.pendingEdits[rowIdx][col] = newValue
    } else if (tab.pendingEdits[rowIdx]) {
      delete tab.pendingEdits[rowIdx][col]
      if (Object.keys(tab.pendingEdits[rowIdx]).length === 0) {
        delete tab.pendingEdits[rowIdx]
      }
    }

    tab.editingCell = null
  }

  function cancelEdit(tab) {
    tab.editingCell = null
  }

  function getAdjacentCell(tab, rowIdx, col, direction = 1) {
    const columns = tab.tableData?.columns || []
    if (!columns.length) return null

    const currentColIndex = columns.indexOf(col)
    if (currentColIndex === -1) return null

    let nextRow = rowIdx
    let nextColIndex = currentColIndex + direction

    if (nextColIndex >= columns.length) {
      nextColIndex = 0
      nextRow += 1
    } else if (nextColIndex < 0) {
      nextColIndex = columns.length - 1
      nextRow -= 1
    }

    if (nextRow < 0 || nextRow >= (tab.tableData?.rows?.length || 0)) {
      return null
    }

    return {
      rowIdx: nextRow,
      col: columns[nextColIndex]
    }
  }

  function handleEditKeydown(event, tab, rowIdx, col) {
    const value = event.target?.value ?? ''

    if (event.key === 'Escape') {
      event.preventDefault()
      cancelEdit(tab)
      return
    }

    if (event.key === 'Enter') {
      event.preventDefault()
      commitEdit(tab, rowIdx, col, value)
      return
    }

    if (event.key === 'Tab') {
      event.preventDefault()
      commitEdit(tab, rowIdx, col, value)
      const nextCell = getAdjacentCell(tab, rowIdx, col, event.shiftKey ? -1 : 1)
      if (nextCell) {
        nextTick(() => startEdit(tab, nextCell.rowIdx, nextCell.col))
      }
    }
  }

  function handleCellBlur(event, tab, rowIdx, col) {
    if (tab.editingCell?.rowIdx !== rowIdx || tab.editingCell?.col !== col) {
      return
    }
    commitEdit(tab, rowIdx, col, event.target?.value ?? '')
  }

  // ── State Queries ──
  function hasPendingEdits(tab) {
    const hasEdits = Object.keys(tab.pendingEdits).some((rowIdx) => (
      Object.keys(tab.pendingEdits[rowIdx] || {}).length > 0
    ))
    const hasNewRows = tab.newRows?.length > 0
    const hasDeleted = tab.deletedRows?.size > 0
    return hasEdits || hasNewRows || hasDeleted
  }

  function revertEdits(tab) {
    tab.pendingEdits = {}
    tab.editingCell = null
    tab.newRows = []
    tab.deletedRows = new Set()
    tab.selectedRows = new Set()
    tab.editingNewCell = null
    tab.selectedNewRows = new Set()
  }

  function getCellValue(tab, row, rowIdx, col) {
    if (tab.pendingEdits[rowIdx] && col in tab.pendingEdits[rowIdx]) {
      return tab.pendingEdits[rowIdx][col]
    }
    return row[col]
  }

  function isRowEdited(tab, rowIdx) {
    return rowIdx in tab.pendingEdits && Object.keys(tab.pendingEdits[rowIdx]).length > 0
  }

  function isEditing(tab, rowIdx, col) {
    return tab.editingCell?.rowIdx === rowIdx && tab.editingCell?.col === col
  }

  function isCellModified(tab, rowIdx, col) {
    return tab.pendingEdits[rowIdx] && col in tab.pendingEdits[rowIdx]
  }

  // ── Edit Mode Toggle ──
  function toggleEditMode(tab) {
    tab.editMode = !tab.editMode
    if (tab.editMode) {
      gmssh.info(t('dataTab.editModeTip'), { duration: 3000 })
    } else {
      tab.editingCell = null
      tab.selectedRows = new Set()
      tab.editingNewCell = null
    }
  }

  // ── Row Selection ──
  function isRowSelected(tab, rowIdx) {
    return tab.selectedRows?.has(rowIdx) || false
  }

  function toggleRowSelect(tab, rowIdx) {
    if (!tab.selectedRows) tab.selectedRows = new Set()
    if (tab.selectedRows.has(rowIdx)) {
      tab.selectedRows.delete(rowIdx)
    } else {
      tab.selectedRows.add(rowIdx)
    }
    tab.selectedRows = new Set(tab.selectedRows)
  }

  function isAllSelected(tab) {
    const rows = tab.tableData?.rows || []
    if (!rows.length) return false
    return tab.selectedRows?.size === rows.length
  }

  function toggleSelectAll(tab, event) {
    const rows = tab.tableData?.rows || []
    if (event.target.checked) {
      tab.selectedRows = new Set(rows.map((_, i) => i))
    } else {
      tab.selectedRows = new Set()
    }
  }

  function hasSelectedRows(tab) {
    return tab.selectedRows?.size > 0
  }

  function isRowDeleted(tab, rowIdx) {
    return tab.deletedRows?.has(rowIdx) || false
  }

  // ── New Row Management ──
  async function addNewRow(tab) {
    if (!tab.newRows) tab.newRows = []

    // Ensure schema is loaded for hint display
    try {
      await deps.ensureDataTabSchema(tab)
    } catch (_) { /* proceed without schema */ }

    const newRow = {}
    ;(tab.tableData?.columns || []).forEach(col => { newRow[col] = null })
    tab.newRows.push(newRow)
  }

  function startNewRowEdit(tab, ni, col) {
    tab.editingNewCell = { rowIdx: ni, col }
  }

  function handleNewRowKeydown(event, tab, ni, col) {
    if (event.key === 'Enter') {
      event.preventDefault()
      commitNewRowEdit(event, tab, ni, col)
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      tab.editingNewCell = null
    }
  }

  function commitNewRowEdit(event, tab, ni, col) {
    const value = event.target?.value ?? ''
    if (!tab.newRows[ni]) return
    tab.newRows[ni][col] = value === '' ? null : value
    tab.editingNewCell = null
  }

  function isNewRowSelected(tab, ni) {
    return tab.selectedNewRows?.has(ni) || false
  }

  function toggleNewRowSelect(tab, ni) {
    if (!tab.selectedNewRows) tab.selectedNewRows = new Set()
    if (tab.selectedNewRows.has(ni)) {
      tab.selectedNewRows.delete(ni)
    } else {
      tab.selectedNewRows.add(ni)
    }
    tab.selectedNewRows = new Set(tab.selectedNewRows)
  }

  // ── Delete ──
  function markDeleteSelected(tab) {
    if (!tab.deletedRows) tab.deletedRows = new Set()
    if (tab.selectedRows) {
      for (const idx of tab.selectedRows) {
        tab.deletedRows.add(idx)
      }
    }
    if (tab.selectedNewRows && tab.newRows) {
      const sorted = [...tab.selectedNewRows].sort((a, b) => b - a)
      sorted.forEach(ni => tab.newRows.splice(ni, 1))
      tab.selectedNewRows = new Set()
    }
    tab.selectedRows = new Set()
    tab.deletedRows = new Set(tab.deletedRows)
  }

  // ── Save with Confirmation ──
  async function confirmSaveEdits(tab) {
    const editCount = Object.keys(tab.pendingEdits).filter(k =>
      Object.keys(tab.pendingEdits[k] || {}).length > 0
    ).length
    const newCount = tab.newRows?.length || 0
    const deleteCount = tab.deletedRows?.size || 0
    const totalChanges = editCount + newCount + deleteCount

    if (totalChanges === 0) return

    const parts = []
    if (editCount > 0) parts.push(t('dataTab.modifiedRows', { n: editCount }))
    if (newCount > 0) parts.push(t('dataTab.newRows', { n: newCount }))
    if (deleteCount > 0) parts.push(t('dataTab.deletedRows', { n: deleteCount }))

    const confirmed = await gmssh.confirm({
      title: t('dataTab.confirmSaveTitle'),
      content: t('dataTab.confirmSaveContent', { detail: parts.join(', ') }),
      positiveText: t('dataTab.confirmSaveYes'),
      negativeText: t('dataTab.confirmSaveNo')
    })
    if (confirmed) {
      await executeAllEdits(tab)
    }
  }

  async function executeAllEdits(tab) {
    await deps.ensureDataTabSchema(tab)
    if (!tab.primaryKeys.length) {
      gmssh.error(t('dataTab.noPKError'))
      return
    }

    const ops = { updates: [], inserts: [], deletes: [] }

    for (const [rowIdxStr, changes] of Object.entries(tab.pendingEdits)) {
      const rowIdx = Number.parseInt(rowIdxStr, 10)
      const row = tab.tableData?.rows?.[rowIdx]
      if (!row || !Object.keys(changes).length) continue

      const pkValues = {}
      tab.primaryKeys.forEach(pk => { pkValues[pk] = row[pk] })
      ops.updates.push({ pkValues, updates: changes })
    }

    if (tab.newRows?.length) {
      for (const newRow of tab.newRows) {
        const values = {}
        Object.entries(newRow).forEach(([k, v]) => {
          if (v != null) values[k] = v
        })
        if (Object.keys(values).length) ops.inserts.push({ values })
      }
    }

    if (tab.deletedRows?.size) {
      for (const rowIdx of tab.deletedRows) {
        const row = tab.tableData?.rows?.[rowIdx]
        if (!row) continue
        const pkValues = {}
        tab.primaryKeys.forEach(pk => { pkValues[pk] = row[pk] })
        ops.deletes.push({ pkValues })
      }
    }

    try {
      const result = await api.batchModify(tab.connId, tab.database, tab.table, ops)

      tab.pendingEdits = {}
      tab.newRows = []
      tab.deletedRows = new Set()
      tab.selectedRows = new Set()
      tab.editingCell = null
      tab.editingNewCell = null
      tab.selectedNewRows = new Set()
      // Clear schema cache so next addNewRow gets fresh auto-increment value
      tab.schemaInfo = null

      const totalSaved = (result?.updated || 0) + (result?.inserted || 0) + (result?.deleted || 0)
      if (totalSaved > 0) {
        gmssh.success(t('dataTab.saveSuccess', { n: totalSaved }))
        deps.refreshTableData(tab)
      }
    } catch (error) {
      gmssh.error(error.message)
    }
  }

  // ── Schema helpers ──
  /**
   * Returns a hint string for new-row display:
   * - auto_increment columns → i18n 'autoIncrement' key
   * - columns with DEFAULT   → i18n 'defaultValue' key
   * - otherwise              → null
   */
  function getColumnHint(tab, col) {
    const colSchema = tab.schemaInfo?.columns?.find(c => c.name === col)
    if (!colSchema) return null
    const extra = (colSchema.extra || '').toLowerCase()
    const defaultVal = (colSchema.defaultValue || '').toLowerCase()
    // MySQL: extra contains 'auto_increment'; PG: defaultValue contains 'nextval'
    if (extra.includes('auto_increment') || extra.includes('nextval') || defaultVal.includes('nextval')) {
      return t('dataTab.autoIncrement')
    }
    if (colSchema.defaultValue) return t('dataTab.defaultValue')
    return null
  }

  return {
    startEdit,
    commitEdit,
    cancelEdit,
    getAdjacentCell,
    handleEditKeydown,
    handleCellBlur,
    hasPendingEdits,
    revertEdits,
    getCellValue,
    isRowEdited,
    isEditing,
    isCellModified,
    toggleEditMode,
    isRowSelected,
    toggleRowSelect,
    isAllSelected,
    toggleSelectAll,
    hasSelectedRows,
    isRowDeleted,
    addNewRow,
    startNewRowEdit,
    handleNewRowKeydown,
    commitNewRowEdit,
    isNewRowSelected,
    toggleNewRowSelect,
    markDeleteSelected,
    confirmSaveEdits,
    executeAllEdits,
    getColumnHint
  }
}
