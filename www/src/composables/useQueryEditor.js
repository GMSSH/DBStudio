/**
 * Composable: Query Tab Editing
 * Handles query result inline editing, row selection, new rows, delete, and batch save.
 */
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'
import { useDatabaseStore } from '@/stores/database'

export function useQueryEditor(t) {
  const store = useDatabaseStore()

  function parseQueryTableName(sql) {
    if (!sql) return null
    const trimmed = sql.trim().replace(/\s+/g, ' ')
    const match = trimmed.match(/^\s*SELECT\s+.+?\s+FROM\s+[`"']?(\w+)[`"']?(?:\s|$|;)/i)
    if (!match) return null
    const upper = trimmed.toUpperCase()
    if (upper.includes(' JOIN ') || upper.includes('(SELECT')) return null
    return match[1]
  }

  function detectQueryEditable(tab) {
    if (tab.queryCanEdit !== null) return
    const tableName = parseQueryTableName(tab.sql)
    if (!tableName) {
      tab.queryCanEdit = false
      tab.queryEditTable = null
      return
    }
    tab.queryEditTable = tableName
    tab.queryCanEdit = true
  }

  async function ensureQueryTabSchema(tab) {
    if (tab.queryPrimaryKeys.length > 0) return
    if (!tab.queryEditTable || !tab.database) return

    try {
      const schema = await store.getTableSchemaFor(tab.queryEditTable, tab.database, tab.connId)
      const schemaData = Array.isArray(schema) ? { columns: schema, primaryKey: [] } : schema
      tab.queryPrimaryKeys = schemaData.primaryKey || []
    } catch (_) {
      tab.queryPrimaryKeys = []
    }
  }

  function toggleQueryEditMode(tab) {
    if (!tab.queryCanEdit) return
    tab.queryEditMode = !tab.queryEditMode
    if (tab.queryEditMode) {
      gmssh.info(t('dataTab.editModeTip'), { duration: 3000 })
    } else {
      tab.queryEditingCell = null
      tab.querySelectedRows = new Set()
      tab.queryEditingNewCell = null
    }
  }

  // ── Cell editing ──
  function isQueryEditing(tab, ri, col) {
    return tab.queryEditingCell?.rowIdx === ri && tab.queryEditingCell?.col === col
  }
  function isQueryCellModified(tab, ri, col) {
    return tab.queryPendingEdits[ri] && col in tab.queryPendingEdits[ri]
  }
  function isQueryRowEdited(tab, ri) {
    return ri in tab.queryPendingEdits && Object.keys(tab.queryPendingEdits[ri]).length > 0
  }
  function isQueryRowSelected(tab, ri) {
    return tab.querySelectedRows?.has(ri) || false
  }
  function isQueryRowDeleted(tab, ri) {
    return tab.queryDeletedRows?.has(ri) || false
  }
  function getQueryCellValue(tab, row, ri, col) {
    if (tab.queryPendingEdits[ri] && col in tab.queryPendingEdits[ri]) {
      return tab.queryPendingEdits[ri][col]
    }
    return row[col]
  }

  function startQueryEdit(tab, ri, col) {
    if (!tab.queryEditMode) return
    tab.queryEditingCell = { rowIdx: ri, col }
  }

  function commitQueryEdit(tab, ri, col, newValue) {
    const row = tab.result?.rows?.[ri]
    const originalValue = row?.[col]
    if (String(newValue) !== String(originalValue ?? '')) {
      if (!tab.queryPendingEdits[ri]) tab.queryPendingEdits[ri] = {}
      tab.queryPendingEdits[ri][col] = newValue
    } else if (tab.queryPendingEdits[ri]) {
      delete tab.queryPendingEdits[ri][col]
      if (Object.keys(tab.queryPendingEdits[ri]).length === 0) delete tab.queryPendingEdits[ri]
    }
    tab.queryEditingCell = null
  }

  function handleQueryEditKeydown(event, tab, ri, col) {
    if (event.key === 'Enter') { event.preventDefault(); commitQueryEdit(tab, ri, col, event.target.value) }
    if (event.key === 'Escape') { event.preventDefault(); tab.queryEditingCell = null }
  }
  function handleQueryCellBlur(event, tab, ri, col) {
    commitQueryEdit(tab, ri, col, event.target.value)
  }

  // ── Row selection ──
  function toggleQueryRowSelect(tab, ri) {
    if (!tab.querySelectedRows) tab.querySelectedRows = new Set()
    if (tab.querySelectedRows.has(ri)) { tab.querySelectedRows.delete(ri) }
    else { tab.querySelectedRows.add(ri) }
    tab.querySelectedRows = new Set(tab.querySelectedRows)
  }
  function toggleQuerySelectAll(tab, event) {
    const rows = tab.result?.rows || []
    if (event.target.checked) {
      tab.querySelectedRows = new Set(rows.map((_, i) => i))
    } else {
      tab.querySelectedRows = new Set()
    }
  }
  function isQueryAllSelected(tab) {
    const rows = tab.result?.rows || []
    if (!rows.length) return false
    return tab.querySelectedRows?.size === rows.length
  }

  // ── New rows ──
  function addQueryNewRow(tab) {
    if (!tab.queryNewRows) tab.queryNewRows = []
    const newRow = {}
    ;(tab.result?.columns || []).forEach(col => { newRow[col] = null })
    tab.queryNewRows.push(newRow)
  }
  function startQueryNewRowEdit(tab, ni, col) {
    tab.queryEditingNewCell = { rowIdx: ni, col }
  }
  function handleQueryNewRowKeydown(event, tab, ni, col) {
    if (event.key === 'Enter') { event.preventDefault(); commitQueryNewRowEdit(event, tab, ni, col) }
    if (event.key === 'Escape') { event.preventDefault(); tab.queryEditingNewCell = null }
  }
  function commitQueryNewRowEdit(event, tab, ni, col) {
    const value = event.target?.value ?? ''
    if (!tab.queryNewRows[ni]) return
    tab.queryNewRows[ni][col] = value === '' ? null : value
    tab.queryEditingNewCell = null
  }
  function toggleQueryNewRowSelect(tab, ni) {
    if (!tab.querySelectedNewRows) tab.querySelectedNewRows = new Set()
    if (tab.querySelectedNewRows.has(ni)) { tab.querySelectedNewRows.delete(ni) }
    else { tab.querySelectedNewRows.add(ni) }
    tab.querySelectedNewRows = new Set(tab.querySelectedNewRows)
  }

  // ── Delete ──
  function markDeleteQuerySelected(tab) {
    if (!tab.queryDeletedRows) tab.queryDeletedRows = new Set()
    if (tab.querySelectedRows) {
      for (const idx of tab.querySelectedRows) { tab.queryDeletedRows.add(idx) }
    }
    if (tab.querySelectedNewRows && tab.queryNewRows) {
      const sorted = [...tab.querySelectedNewRows].sort((a, b) => b - a)
      sorted.forEach(ni => tab.queryNewRows.splice(ni, 1))
      tab.querySelectedNewRows = new Set()
    }
    tab.querySelectedRows = new Set()
    tab.queryDeletedRows = new Set(tab.queryDeletedRows)
  }

  // ── Pending state ──
  function hasQueryPendingEdits(tab) {
    const hasEdits = Object.keys(tab.queryPendingEdits).some(k =>
      Object.keys(tab.queryPendingEdits[k] || {}).length > 0
    )
    return hasEdits || tab.queryNewRows?.length > 0 || tab.queryDeletedRows?.size > 0
  }
  function countQueryEdits(tab) {
    const editCount = Object.keys(tab.queryPendingEdits).filter(k =>
      Object.keys(tab.queryPendingEdits[k] || {}).length > 0
    ).length
    return editCount + (tab.queryNewRows?.length || 0) + (tab.queryDeletedRows?.size || 0)
  }
  function revertQueryEdits(tab) {
    tab.queryPendingEdits = {}
    tab.queryEditingCell = null
    tab.queryNewRows = []
    tab.queryDeletedRows = new Set()
    tab.querySelectedRows = new Set()
    tab.queryEditingNewCell = null
    tab.querySelectedNewRows = new Set()
  }

  // ── Save with transaction ──
  async function confirmSaveQueryEdits(tab) {
    const editCount = Object.keys(tab.queryPendingEdits).filter(k =>
      Object.keys(tab.queryPendingEdits[k] || {}).length > 0
    ).length
    const newCount = tab.queryNewRows?.length || 0
    const deleteCount = tab.queryDeletedRows?.size || 0
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
      await executeQueryEdits(tab)
    }
  }

  async function executeQueryEdits(tab) {
    await ensureQueryTabSchema(tab)
    if (!tab.queryPrimaryKeys.length) {
      gmssh.error(t('dataTab.noPKError'))
      return
    }

    const tableName = tab.queryEditTable
    const ops = { updates: [], inserts: [], deletes: [] }

    for (const [rowIdxStr, changes] of Object.entries(tab.queryPendingEdits)) {
      const rowIdx = Number.parseInt(rowIdxStr, 10)
      const row = tab.result?.rows?.[rowIdx]
      if (!row || !Object.keys(changes).length) continue
      const pkValues = {}
      tab.queryPrimaryKeys.forEach(pk => { pkValues[pk] = row[pk] })
      ops.updates.push({ pkValues, updates: changes })
    }

    if (tab.queryNewRows?.length) {
      for (const newRow of tab.queryNewRows) {
        const values = {}
        Object.entries(newRow).forEach(([k, v]) => { if (v != null) values[k] = v })
        if (Object.keys(values).length) ops.inserts.push({ values })
      }
    }

    if (tab.queryDeletedRows?.size) {
      for (const rowIdx of tab.queryDeletedRows) {
        const row = tab.result?.rows?.[rowIdx]
        if (!row) continue
        const pkValues = {}
        tab.queryPrimaryKeys.forEach(pk => { pkValues[pk] = row[pk] })
        ops.deletes.push({ pkValues })
      }
    }

    try {
      const result = await api.batchModify(tab.connId, tab.database, tableName, ops)
      revertQueryEdits(tab)
      const totalSaved = (result?.updated || 0) + (result?.inserted || 0) + (result?.deleted || 0)
      if (totalSaved > 0) {
        gmssh.success(t('dataTab.saveSuccess', { n: totalSaved }))
      }
    } catch (error) {
      gmssh.error(error.message)
    }
  }

  return {
    parseQueryTableName,
    detectQueryEditable,
    ensureQueryTabSchema,
    toggleQueryEditMode,
    isQueryEditing,
    isQueryCellModified,
    isQueryRowEdited,
    isQueryRowSelected,
    isQueryRowDeleted,
    getQueryCellValue,
    startQueryEdit,
    commitQueryEdit,
    handleQueryEditKeydown,
    handleQueryCellBlur,
    toggleQueryRowSelect,
    toggleQuerySelectAll,
    isQueryAllSelected,
    addQueryNewRow,
    startQueryNewRowEdit,
    handleQueryNewRowKeydown,
    commitQueryNewRowEdit,
    toggleQueryNewRowSelect,
    markDeleteQuerySelected,
    hasQueryPendingEdits,
    countQueryEdits,
    revertQueryEdits,
    confirmSaveQueryEdits,
    executeQueryEdits
  }
}
