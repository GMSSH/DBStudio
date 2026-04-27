<template>
  <div ref="workAreaRef" class="work-area">
    <!-- Unified Tab System -->
    <n-tabs 
      v-if="workspaceTabs.length > 0"
      v-model:value="activeTabId" 
      type="card" 
      closable
      :on-close="closeTab"
      class="workspace-tabs"
    >
      <template #suffix>
        <n-popover
          v-model:show="tabSwitcherVisible"
          trigger="click"
          placement="bottom-end"
          class="tab-switcher-popover"
        >
          <template #trigger>
            <button class="tab-switcher-trigger" type="button" :aria-label="t('tabSwitcher.trigger')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M7 10l5 5 5-5" />
              </svg>
            </button>
          </template>

          <div class="tab-switcher-panel">
            <div class="tab-switcher-search">
              <n-input
                v-model:value="tabSearch"
                size="small"
                clearable
                :placeholder="t('tabSwitcher.searchPlaceholder')"
              >
                <template #prefix>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="11" cy="11" r="7" />
                    <path d="M20 20l-3.5-3.5" />
                  </svg>
                </template>
              </n-input>
            </div>

            <div class="tab-switcher-section-title">
              {{ t('tabSwitcher.openTabs', { n: workspaceTabs.length }) }}
            </div>

            <div v-if="filteredWorkspaceTabs.length > 0" class="tab-switcher-list">
              <div
                v-for="tab in filteredWorkspaceTabs"
                :key="`switcher-${tab.id}`"
                role="button"
                tabindex="0"
                :class="['tab-switcher-item', activeTabId === tab.id ? 'tab-switcher-item--active' : '']"
                @click="selectWorkspaceTab(tab.id)"
                @keydown.enter.prevent="selectWorkspaceTab(tab.id)"
                @keydown.space.prevent="selectWorkspaceTab(tab.id)"
              >
                <span class="tab-switcher-item__icon">
                  <svg v-if="tab.type === 'query'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="4 17 10 11 4 5" />
                    <line x1="12" y1="19" x2="20" y2="19" />
                  </svg>
                  <svg v-else-if="tab.type === 'overview'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <ellipse cx="12" cy="5" rx="8" ry="3" />
                    <path d="M4 5v6c0 1.66 3.58 3 8 3s8-1.34 8-3V5" />
                    <path d="M4 11v6c0 1.66 3.58 3 8 3s8-1.34 8-3v-6" />
                  </svg>
                  <svg v-else-if="tab.type === 'structure'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5z" />
                    <path d="M2 17l10 5 10-5" />
                    <path d="M2 12l10 5 10-5" />
                  </svg>
                  <svg v-else-if="tab.type === 'transfer'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="17 8 12 3 7 8" />
                    <line x1="12" y1="3" x2="12" y2="15" />
                  </svg>
                  <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="3" width="18" height="18" rx="2" />
                    <path d="M3 9h18M3 15h18M9 3v18M15 3v18" />
                  </svg>
                </span>

                <span class="tab-switcher-item__content">
                  <span class="tab-switcher-item__title">{{ getTabPlainTitle(tab) }}</span>
                  <span class="tab-switcher-item__meta">
                    {{ getTabPlainSubtitle(tab) }}
                  </span>
                </span>

                <button
                  type="button"
                  class="tab-switcher-item__close"
                  :aria-label="t('tabSwitcher.closeTab')"
                  @click.stop="closeTab(tab.id, tab.connId)"
                >
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 6L6 18M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>

            <n-empty
              v-else
              size="small"
              :description="t('tabSwitcher.empty')"
              class="tab-switcher-empty"
            />
          </div>
        </n-popover>
      </template>

      <n-tab-pane 
        v-for="tab in workspaceTabs" 
        :key="tab.id" 
        :name="tab.id" 
        display-directive="show"
        :tab="getTabLabel(tab)"
      >
        <!-- Query Tab (Navicat-style layout) -->
        <template v-if="tab.type === 'query'">
          <div class="query-layout">
            <!-- Compact Toolbar -->
            <div class="query-toolbar">
              <div class="query-toolbar-left">
                <n-select
                  v-model:value="tab.database"
                  :options="databaseOptions"
                  :render-label="renderDbLabel"
                  :placeholder="t('query.selectDb')"
                  style="width: 160px;"
                  size="small"
                  @update:value="() => handleDatabaseChange(tab)"
                />
                <!-- Schema selector (PostgreSQL only) -->
                <n-select
                  v-if="isPostgreSQL"
                  v-model:value="tab.schema"
                  :options="schemaOptions"
                  :placeholder="t('query.selectSchema')"
                  style="width: 140px; margin-left: 8px;"
                  size="small"
                  :disabled="!tab.database"
                  :loading="loadingSchemas"
                />
                <div class="query-toolbar-divider"></div>
                <n-button 
                  size="small" 
                  type="primary" 
                  @click="executeTabSQL(tab)" 
                  :loading="tab.executing" 
                  :disabled="!tab.database"
                  class="run-btn"
                >
                  <template #icon>
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                      <polygon points="5,3 19,12 5,21"/>
                    </svg>
                  </template>
                  {{ t('query.run') }}
                </n-button>
                <n-button v-if="tab.executing" size="small" type="error" @click="stopTabExecution(tab)">
                  {{ t('query.stop') }}
                </n-button>
                <n-button size="small" quaternary @click="openHistoryPanel(tab.id)">
                  {{ t('query.history') }}
                </n-button>
              </div>
              <div class="query-toolbar-right">
                <span v-if="!tab.database" class="query-hint">{{ t('query.noDbHint') }}</span>
                <span v-if="tab.result?.success" class="query-status-ok">
                  <span class="status-led"></span>
                  {{ t('query.rows', { n: tab.result.rowCount }) }}&nbsp;&nbsp;·&nbsp;&nbsp;{{ formatDuration(tab.result.executionTime) }}
                </span>
              </div>
            </div>

            <!-- Editor Area (fills available space) -->
            <div class="query-editor-area" :style="{ flex: tab.result ? `0 0 ${tab.editorHeight || 55}%` : '1' }">
              <SQLEditor 
                :ref="el => setTabEditorRef(tab.id, el)"
                v-model="tab.sql" 
                @execute="() => executeTabSQL(tab)"
                :db-type="store.connectionConfig?.dbType || 'mysql'"
                :completion-schema="getCompletionSchema(tab.database)"
                :full-height="true"
              />
            </div>

            <!-- Resizable Splitter - only show when there's a result panel -->
            <div 
              v-if="tab.result"
              class="query-splitter"
              @mousedown="(e) => startSplitterDrag(e, tab)"
            ></div>

            <!-- Results Panel - show on any result (success or error) -->
            <div v-if="tab.result" class="query-results" :style="{ flex: `0 0 ${100 - (tab.editorHeight || 55)}%` }">

              <!-- Error State -->
              <n-alert v-if="tab.result.error" type="error" :bordered="false" style="margin: 8px 0;">
                {{ tab.result.error }}
              </n-alert>

              <!-- Premium Table (with query editing support) -->
              <template v-if="tab.result.rows && tab.result.rows.length > 0">
                <!-- Pending edits bar for query results -->
                <div v-if="hasQueryPendingEdits(tab)" class="edit-action-bar" style="flex-shrink:0">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="rgba(255,214,0,0.9)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                  <span>{{ t('dataTab.pendingEdits', { n: countQueryEdits(tab) }) }}</span>
                  <n-button size="tiny" type="primary" @click="confirmSaveQueryEdits(tab)">
                    <template #icon><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg></template>
                    {{ t('dataTab.saveEdits') }}
                  </n-button>
                  <n-button size="tiny" @click="revertQueryEdits(tab)">
                    <template #icon><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg></template>
                    {{ t('dataTab.revert') }}
                  </n-button>
                </div>
                <div class="premium-table-wrap">
                  <table class="premium-table">
                    <thead>
                      <tr>
                        <th v-if="tab.queryEditMode" class="gm-pth ptd-checkbox" style="width:36px;">
                          <input type="checkbox" @change="toggleQuerySelectAll(tab, $event)" :checked="isQueryAllSelected(tab)" />
                        </th>
                        <th
                          v-for="col in tab.result.columns"
                          :key="col"
                          class="gm-pth"
                        >
                          {{ col }}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(row, ri) in tab.result.rows"
                        :key="ri"
                        :class="['gm-ptr', ri % 2 === 1 ? 'ptr-even' : '', isQueryRowEdited(tab, ri) ? 'ptr-edited' : '', isQueryRowSelected(tab, ri) ? 'ptr-selected' : '', isQueryRowDeleted(tab, ri) ? 'ptr-deleted' : '']" 
                      >
                        <td v-if="tab.queryEditMode" class="gm-ptd ptd-checkbox">
                          <input type="checkbox" :checked="isQueryRowSelected(tab, ri)" @change="toggleQueryRowSelect(tab, ri)" />
                        </td>
                        <td
                          v-for="col in tab.result.columns"
                          :key="col"
                          class="gm-ptd"
                          :class="{ 'ptd-editing': isQueryEditing(tab, ri, col), 'ptd-modified': isQueryCellModified(tab, ri, col) }"
                          :title="row[col] == null ? '' : String(row[col])"
                          @dblclick="tab.queryEditMode && startQueryEdit(tab, ri, col)"
                        >
                          <input
                            v-if="isQueryEditing(tab, ri, col)"
                            class="cell-edit-input"
                            :value="getQueryCellValue(tab, row, ri, col)"
                            @keydown="handleQueryEditKeydown($event, tab, ri, col)"
                            @blur="handleQueryCellBlur($event, tab, ri, col)"
                            v-auto-focus
                          />
                          <template v-else>
                            <span v-if="getQueryCellValue(tab, row, ri, col) == null" class="null-badge">null</span>
                            <span v-else class="cell-content">{{ getQueryCellValue(tab, row, ri, col) }}</span>
                          </template>
                        </td>
                      </tr>
                      <!-- New rows in query result -->
                      <tr v-for="(newRow, ni) in (tab.queryNewRows || [])" :key="'qnew-' + ni" class="gm-ptr ptr-new">
                        <td v-if="tab.queryEditMode" class="gm-ptd ptd-checkbox">
                          <input type="checkbox" :checked="tab.querySelectedNewRows?.has(ni)" @change="toggleQueryNewRowSelect(tab, ni)" />
                        </td>
                        <td v-for="col in tab.result.columns" :key="col" class="gm-ptd" @dblclick="startQueryNewRowEdit(tab, ni, col)">
                          <input
                            v-if="tab.queryEditingNewCell?.rowIdx === ni && tab.queryEditingNewCell?.col === col"
                            class="cell-edit-input"
                            :value="newRow[col] ?? ''"
                            @keydown="handleQueryNewRowKeydown($event, tab, ni, col)"
                            @blur="commitQueryNewRowEdit($event, tab, ni, col)"
                            v-auto-focus
                          />
                          <template v-else>
                            <span v-if="newRow[col] == null || newRow[col] === ''" class="null-badge">null</span>
                            <span v-else class="cell-content">{{ newRow[col] }}</span>
                          </template>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <!-- Query result bottom bar -->
                <div class="data-bottom-bar" style="flex-shrink:0;">
                  <div class="bottom-bar-left">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <button class="bar-icon-btn" :class="{ 'bar-icon-btn--active': tab.queryEditMode }" @click="toggleQueryEditMode(tab)" :disabled="!tab.queryCanEdit">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                        </button>
                      </template>
                      {{ tab.queryCanEdit === false ? t('query.editNotSupported') : (tab.queryEditMode ? t('dataTab.exitEdit') : t('dataTab.enterEdit')) }}
                    </n-tooltip>
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <button class="bar-icon-btn" :disabled="!tab.queryEditMode" @click="addQueryNewRow(tab)">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                        </button>
                      </template>
                      {{ t('dataTab.addRow') }}
                    </n-tooltip>
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <button class="bar-icon-btn bar-icon-btn--danger" :disabled="!tab.queryEditMode || !tab.querySelectedRows?.size" @click="markDeleteQuerySelected(tab)">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        </button>
                      </template>
                      {{ t('dataTab.deleteRows') }}
                    </n-tooltip>
                  </div>
                  <div class="bottom-bar-right">
                    <span class="bar-stats" v-if="tab.queryEditTable">
                      <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="#34C759" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                      {{ tab.queryEditTable }}
                    </span>
                    <span class="bar-stats" v-else-if="tab.queryCanEdit === false" style="color:#FF6B6B">
                      {{ t('query.editNotSupported') }}
                    </span>
                    <span class="bar-divider"></span>
                    <span class="bar-stats">{{ tab.result.rows?.length || 0 }} {{ t('dataTab.rowsUnit') }}</span>
                  </div>
                </div>
              </template>

              <!-- Premium Empty State -->
              <div v-else-if="tab.result.success" class="premium-empty">
                <svg class="empty-svg" viewBox="0 0 120 100" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <ellipse cx="60" cy="82" rx="38" ry="8" fill="rgba(94,159,248,0.08)"/>
                  <circle cx="60" cy="46" r="28" fill="rgba(94,159,248,0.06)" stroke="rgba(94,159,248,0.15)" stroke-width="1.5"/>
                  <circle cx="60" cy="46" r="18" fill="rgba(94,159,248,0.08)" stroke="rgba(94,159,248,0.2)" stroke-width="1"/>
                  <!-- magnifier handle -->
                  <line x1="73" y1="59" x2="84" y2="72" stroke="rgba(94,159,248,0.4)" stroke-width="3" stroke-linecap="round"/>
                  <!-- magnifier glass -->
                  <circle cx="60" cy="46" r="18" stroke="rgba(94,159,248,0.35)" stroke-width="2" fill="none"/>
                  <!-- inner shine -->
                  <path d="M50 38 Q54 34 60 35" stroke="rgba(255,255,255,0.25)" stroke-width="1.5" stroke-linecap="round" fill="none"/>
                  <!-- small dots -->
                  <circle cx="51" cy="47" r="1.5" fill="rgba(94,159,248,0.5)"/>
                  <circle cx="60" cy="47" r="1.5" fill="rgba(94,159,248,0.5)"/>
                  <circle cx="69" cy="47" r="1.5" fill="rgba(94,159,248,0.5)"/>
                </svg>
                <div class="empty-title">{{ t('query.noDataTitle') }}</div>
                <div class="empty-sub">{{ t('query.noRowsReturned') }}</div>
              </div>

            </div>
          </div>
        </template>

        <!-- Database Overview Tab -->
        <template v-else-if="tab.type === 'overview'">
          <div class="overview-tab-pane">
            <Suspense>
              <DatabaseOverview
                :database="tab.database"
                :conn-id="tab.connId"
                :cache="tab.overviewCache"
                @refresh="() => refreshOverviewTab(tab)"
                @open-data="(object) => openTableDataTab(tab.database, object?.name, tab.connId)"
                @open-structure="(object) => openStructureTab(tab.database, object?.name, tab.connId)"
                @open-query="(object) => openQueryTabWithContext(tab.database, object?.name || null, tab.connId)"
                @open-export="(object) => openExportModal(object?.name, tab.database, { connId: tab.connId })"
                @open-import="(object) => openImportModal(object?.name, tab.database, { connId: tab.connId })"
              />
            </Suspense>
          </div>
        </template>

        <!-- Table Data Tab -->
        <template v-else-if="tab.type === 'data'">
          <div class="data-tab-layout">

            <!-- Pending edits action bar (shown when there are unsaved changes) -->
            <div v-if="hasPendingEdits(tab)" class="edit-action-bar">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="rgba(255,214,0,0.9)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              <span>{{ t('dataTab.pendingEdits', { n: Object.keys(tab.pendingEdits).length }) }}</span>
              <n-button size="tiny" type="primary" @click="confirmSaveEdits(tab)">
                <template #icon><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg></template>
                {{ t('dataTab.saveEdits') }}
              </n-button>
              <n-button size="tiny" @click="revertEdits(tab)">
                <template #icon><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg></template>
                {{ t('dataTab.revert') }}
              </n-button>
            </div>

            <!-- Data Table with inline editing -->
            <n-spin :show="tab.loading" class="data-spin-wrap">
              <template v-if="tab.tableData && tab.tableData.columns && tab.tableData.columns.length > 0">

                <!-- TABLE VIEW (grid mode) -->
                <div v-if="!tab.formViewMode" class="premium-table-wrap">
                  <table class="premium-table">
                    <thead>
                      <tr>
                        <!-- Checkbox column (only in edit mode) -->
                        <th v-if="tab.editMode" class="gm-pth ptd-checkbox" style="width:36px;">
                          <input type="checkbox" @change="toggleSelectAll(tab, $event)" :checked="isAllSelected(tab)" />
                        </th>
                        <th
                          v-for="col in tab.tableData.columns"
                          :key="col"
                          class="gm-pth pth-sortable"
                          @click="toggleSort(tab, col)"
                        >
                          <span class="th-content">
                            {{ col }}
                            <span class="sort-indicator">
                              <svg v-if="tab.sortCol === col && tab.sortDir === 'asc'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="var(--ref-color-brand-5)" stroke-width="2.5"><polyline points="18 15 12 9 6 15"/></svg>
                              <svg v-else-if="tab.sortCol === col && tab.sortDir === 'desc'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="var(--ref-color-brand-5)" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg>
                              <svg v-else width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.2)" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg>
                            </span>
                          </span>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <!-- Existing rows -->
                      <tr
                        v-for="(row, ri) in tab.tableData.rows"
                        :key="ri"
                        :class="['gm-ptr', ri % 2 === 1 ? 'ptr-even' : '', tab.activeRowIndex === ri ? 'ptr-active' : '', isRowEdited(tab, ri) ? 'ptr-edited' : '', isRowSelected(tab, ri) ? 'ptr-selected' : '', isRowDeleted(tab, ri) ? 'ptr-deleted' : '']"
                        @click="tab.activeRowIndex = ri"
                      >
                        <!-- Checkbox -->
                        <td v-if="tab.editMode" class="gm-ptd ptd-checkbox">
                          <input type="checkbox" :checked="isRowSelected(tab, ri)" @change="toggleRowSelect(tab, ri)" />
                        </td>
                        <td
                          v-for="col in tab.tableData.columns"
                          :key="col"
                          class="gm-ptd"
                          :class="{ 'ptd-editing': isEditing(tab, ri, col), 'ptd-modified': isCellModified(tab, ri, col) }"
                          @dblclick="tab.editMode && startEdit(tab, ri, col)"
                        >
                          <!-- Editing state -->
                          <input
                            v-if="isEditing(tab, ri, col)"
                            class="cell-edit-input"
                            :value="getCellValue(tab, row, ri, col)"
                            @keydown="handleEditKeydown($event, tab, ri, col)"
                            @blur="handleCellBlur($event, tab, ri, col)"
                            ref="cellEditInput"
                            v-auto-focus
                          />
                          <!-- Display state -->
                          <template v-else>
                            <span v-if="getCellValue(tab, row, ri, col) == null" class="null-badge">null</span>
                            <span v-else class="cell-content" :title="String(getCellValue(tab, row, ri, col))">{{ getCellValue(tab, row, ri, col) }}</span>
                          </template>
                        </td>
                      </tr>
                      <!-- New rows (pending insert) -->
                      <tr
                        v-for="(newRow, ni) in (tab.newRows || [])"
                        :key="'new-' + ni"
                        class="gm-ptr ptr-new"
                      >
                        <td v-if="tab.editMode" class="gm-ptd ptd-checkbox">
                          <input type="checkbox" :checked="isNewRowSelected(tab, ni)" @change="toggleNewRowSelect(tab, ni)" />
                        </td>
                        <td
                          v-for="col in tab.tableData.columns"
                          :key="col"
                          class="gm-ptd"
                          @dblclick="startNewRowEdit(tab, ni, col)"
                        >
                          <input
                            v-if="tab.editingNewCell?.rowIdx === ni && tab.editingNewCell?.col === col"
                            class="cell-edit-input"
                            :value="newRow[col] ?? ''"
                            @keydown="handleNewRowKeydown($event, tab, ni, col)"
                            @blur="commitNewRowEdit($event, tab, ni, col)"
                            v-auto-focus
                          />
                          <template v-else>
                            <span v-if="(newRow[col] == null || newRow[col] === '') && getColumnHint(tab, col)" class="hint-badge">{{ getColumnHint(tab, col) }}</span>
                            <span v-else-if="newRow[col] == null || newRow[col] === ''" class="null-badge">null</span>
                            <span v-else class="cell-content">{{ newRow[col] }}</span>
                          </template>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <!-- FORM VIEW (single record mode) -->
                <div v-else class="form-view-wrap">
                  <template v-if="tab.activeRowIndex != null && tab.tableData.rows[tab.activeRowIndex]">
                    <div class="form-view-nav">
                      <button class="bar-icon-btn bar-icon-btn--sm" :disabled="tab.activeRowIndex <= 0" @click="tab.activeRowIndex--">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
                      </button>
                      <span class="form-view-counter">{{ t('dataTab.recordOf', { current: tab.activeRowIndex + 1, total: tab.tableData.rows.length }) }}</span>
                      <button class="bar-icon-btn bar-icon-btn--sm" :disabled="tab.activeRowIndex >= tab.tableData.rows.length - 1" @click="tab.activeRowIndex++">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
                      </button>
                    </div>
                    <div class="form-view-fields">
                      <div
                        v-for="col in tab.tableData.columns"
                        :key="col"
                        class="form-view-field"
                        :class="{ 'form-view-field--modified': isCellModified(tab, tab.activeRowIndex, col) }"
                      >
                        <label class="form-view-label">{{ col }}</label>
                        <div class="form-view-value">
                          <!-- Editable mode -->
                          <template v-if="tab.editMode">
                            <textarea
                              class="form-view-input"
                              :value="getFormCellValue(tab, tab.activeRowIndex, col)"
                              @blur="commitFormEdit($event, tab, tab.activeRowIndex, col)"
                              @keydown.ctrl.enter="$event.target.blur()"
                              @keydown.meta.enter="$event.target.blur()"
                              rows="1"
                              @input="autoResizeTextarea($event)"
                            />
                          </template>
                          <!-- Read-only mode -->
                          <template v-else>
                            <span v-if="getCellValue(tab, tab.tableData.rows[tab.activeRowIndex], tab.activeRowIndex, col) == null" class="null-badge">null</span>
                            <pre v-else class="form-view-pre">{{ String(getCellValue(tab, tab.tableData.rows[tab.activeRowIndex], tab.activeRowIndex, col)) }}</pre>
                          </template>
                        </div>
                      </div>
                    </div>
                  </template>
                  <div v-else class="form-view-empty">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.15)" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/></svg>
                    <span>{{ t('dataTab.noRowSelected') }}</span>
                  </div>
                </div>

              </template>
              <n-empty v-else-if="!tab.loading" :description="t('dataTab.noData')" style="margin-top: 48px;" />
            </n-spin>

            <!-- Bottom Action Bar -->
            <div class="data-bottom-bar">
              <div class="bottom-bar-left">
                <!-- Edit mode toggle -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button
                      class="bar-icon-btn"
                      :class="{ 'bar-icon-btn--active': tab.editMode }"
                      @click="toggleEditMode(tab)"
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    </button>
                  </template>
                  {{ tab.editMode ? t('dataTab.exitEdit') : t('dataTab.enterEdit') }}
                </n-tooltip>

                <!-- Add row -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button class="bar-icon-btn" :disabled="!tab.editMode" @click="handleAddNewRow(tab)">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                    </button>
                  </template>
                  {{ t('dataTab.addRow') }}
                </n-tooltip>

                <!-- Delete selected rows -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button class="bar-icon-btn bar-icon-btn--danger" :disabled="!tab.editMode || !hasSelectedRows(tab)" @click="markDeleteSelected(tab)">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                    </button>
                  </template>
                  {{ t('dataTab.deleteRows') }}
                </n-tooltip>

                <span class="bar-divider"></span>

                <!-- Refresh -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button class="bar-icon-btn" :disabled="tab.loading" @click="refreshTableData(tab)">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                    </button>
                  </template>
                  {{ t('dataTab.refresh') }}
                </n-tooltip>

                <!-- Export -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button class="bar-icon-btn" @click="openExportModal(tab.table, tab.database, { rows: tab.tableData?.rows || [], columns: tab.tableData?.columns || [] })">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                    </button>
                  </template>
                  {{ t('ie.export') }}
                </n-tooltip>

                <span class="bar-divider"></span>

                <!-- Form/Table view toggle -->
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button
                      class="bar-icon-btn"
                      :class="{ 'bar-icon-btn--active': tab.formViewMode }"
                      @click="tab.formViewMode = !tab.formViewMode"
                    >
                      <!-- Form icon when in table mode, Grid icon when in form mode -->
                      <svg v-if="!tab.formViewMode" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="3" y1="15" x2="21" y2="15"/><line x1="9" y1="3" x2="9" y2="21"/></svg>
                      <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
                    </button>
                  </template>
                  {{ tab.formViewMode ? t('dataTab.tableView') : t('dataTab.formView') }}
                </n-tooltip>
              </div>

              <div class="bottom-bar-right">
                <!-- Pagination -->
                <span class="bar-stats">{{ tab.tableData?.total || 0 }} {{ t('dataTab.rowsUnit') }}</span>
                <span class="bar-divider"></span>
                <button class="bar-icon-btn bar-icon-btn--sm" :disabled="tab.page <= 1 || tab.loading" @click="() => { tab.page--; refreshTableData(tab) }">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
                </button>
                <input
                  class="page-jump-input"
                  type="text"
                  :key="'pj-' + tab.id + '-' + tab.page"
                  :value="tab.page"
                  @keydown.enter="$event.target.blur()"
                  @blur="jumpToPage(tab, $event.target.value)"
                />
                <span class="bar-page-info">/ {{ Math.max(1, Math.ceil((tab.tableData?.total || 0) / tab.pageSize)) }}</span>
                <button class="bar-icon-btn bar-icon-btn--sm" :disabled="(tab.page * tab.pageSize) >= (tab.tableData?.total || 0) || tab.loading" @click="() => { tab.page++; refreshTableData(tab) }">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
                </button>
                <span class="bar-divider"></span>
                <span class="bar-stats">{{ tab.pageSize }}/{{ t('dataTab.page') }}</span>
              </div>
            </div>

          </div>
        </template>

        <!-- Table Structure Tab -->
        <template v-else-if="tab.type === 'structure'">
          <div class="structure-tab-layout">
            <div class="structure-content-shell">
              <n-spin :show="tab.loading" class="data-spin-wrap">
                <div v-if="tab.schema" class="premium-table-wrap">
                  <table class="premium-table">
                    <thead>
                      <tr>
                        <th class="gm-pth">{{ t('schema.colName') }}</th>
                        <th class="gm-pth">{{ t('schema.colType') }}</th>
                        <th class="gm-pth" style="width:72px;text-align:center">{{ t('schema.colNullable') }}</th>
                        <th class="gm-pth">{{ t('schema.colDefault') }}</th>
                        <th class="gm-pth" style="width:60px;text-align:center">{{ t('schema.colPrimaryKey') }}</th>
                        <th class="gm-pth">{{ t('schema.colComment') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(col, ri) in tab.schema"
                        :key="col.name"
                        :class="['gm-ptr', ri % 2 === 1 ? 'ptr-even' : '']"
                      >
                        <td class="gm-ptd ptd-name">{{ col.name }}</td>
                        <td class="gm-ptd ptd-mono">{{ col.type }}</td>
                        <td class="gm-ptd ptd-center">
                          <span v-if="col.nullable" class="schema-badge schema-badge--yes">{{ t('schema.colNullableYes') }}</span>
                          <span v-else class="schema-badge schema-badge--no">{{ t('schema.colNullableNo') }}</span>
                        </td>
                        <td class="gm-ptd ptd-mono">{{ col.defaultValue ?? '' }}</td>
                        <td class="gm-ptd ptd-center">
                          <svg v-if="col.isPrimaryKey" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="var(--ref-color-yellow-6)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                        </td>
                        <td class="gm-ptd ptd-comment">{{ col.comment || '' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <n-empty v-else :description="t('schema.noData')" />
              </n-spin>

              <!-- Inspector Sidebar (same as old data tab inspector) -->
              <aside v-if="tab.showInspector" class="table-inspector">
                <div class="inspector-card">
                  <div class="inspector-section-title">{{ t('dataTab.basicInfo') }}</div>
                  <div class="inspector-kv-list">
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.objectLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.table) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('overview.databaseLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.database) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.rowsLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.metaInfo?.rowCount) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.columnsLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.schema?.length) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.engineLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.metaInfo?.engine) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.sizeLabel') }}</span>
                      <strong>{{ formatSize(tab.metaInfo?.size) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.createdAtLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.metaInfo?.createdAt) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.updatedAtLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.metaInfo?.updatedAt) }}</strong>
                    </div>
                    <div class="inspector-kv">
                      <span>{{ t('dataTab.commentLabel') }}</span>
                      <strong>{{ formatInspectorValue(tab.metaInfo?.comment) }}</strong>
                    </div>
                  </div>
                </div>

                <div class="inspector-card">
                  <div class="inspector-section-head">
                    <div class="inspector-section-title">{{ t('dataTab.ddlTitle') }}</div>
                    <n-button size="tiny" quaternary :loading="tab.ddlLoading" @click="loadTableDDL(tab, true)">
                      {{ t('dataTab.refreshDDL') }}
                    </n-button>
                  </div>
                  <pre class="inspector-ddl">{{ tab.ddlText || t('dataTab.ddlEmpty') }}</pre>
                  <div class="inspector-actions">
                    <n-button size="small" quaternary :disabled="!tab.ddlText" @click="copyInspectorDDL(tab)">
                      {{ t('dataTab.copyDDL') }}
                    </n-button>
                  </div>
                </div>
              </aside>
            </div>

            <!-- Bottom bar for structure tab -->
            <div class="data-bottom-bar">
              <div class="bottom-bar-left">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button class="bar-icon-btn" :disabled="tab.loading" @click="refreshTableStructure(tab)">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                    </button>
                  </template>
                  {{ t('dataTab.refresh') }}
                </n-tooltip>

                <n-tooltip trigger="hover">
                  <template #trigger>
                    <button
                      class="bar-icon-btn"
                      :class="{ 'bar-icon-btn--active': tab.showInspector }"
                      @click="toggleInspector(tab)"
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                    </button>
                  </template>
                  {{ t('dataTab.inspector') }}
                </n-tooltip>
              </div>
              <div class="bottom-bar-right">
                <span class="bar-stats">{{ tab.schema?.length || 0 }} {{ t('dataTab.columnsLabel') }}</span>
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="tab.type === 'transfer'">
          <Suspense>
            <TransferCenter
              :conn-id="tab.connId"
              :initial-database="tab.database"
              :initial-action="tab.action"
              :auto-launch="tab.autoLaunch"
              :launch-key="tab.launchKey"
              @auto-launch-consumed="tab.autoLaunch = false"
            />
          </Suspense>
        </template>
      </n-tab-pane>
    </n-tabs>
    
    <!-- Welcome message when no tabs -->
    <div v-else class="welcome-state">
      <div class="welcome-icon">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" opacity="0.3">
          <ellipse cx="12" cy="6" rx="8" ry="3"/>
          <path d="M4 6v12c0 1.66 3.58 3 8 3s8-1.34 8-3V6"/>
          <path d="M4 12c0 1.66 3.58 3 8 3s8-1.34 8-3"/>
        </svg>
      </div>
      <div class="welcome-title">{{ t('welcome.title') }}</div>
      <div class="welcome-hint">{{ t('welcome.hint') }}</div>
    </div>

    <!-- Import/Export Modal (global, above tabs) -->
    <Suspense>
      <ImportExport
        v-if="ieVisible"
        v-model="ieVisible"
        :mode="ieMode"
        :table-name="ieTableName"
        :database="ieDatabase"
        :conn-id="ieConnId"
        :page-rows="iePageRows"
        :page-columns="iePageColumns"
        @imported="() => { const dt = (getWorkspace(ieConnId)?.tabs || workspaceTabs).find(t => t.type === 'data' && t.table === ieTableName); if (dt) refreshTableData(dt) }"
      />
    </Suspense>

    <Suspense>
      <SQLHistoryPanel
        v-if="historyVisible"
        v-model="historyVisible"
        @select="handleHistorySelect"
      />
    </Suspense>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onUnmounted, nextTick, h, defineAsyncComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTooltip } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

// Custom directive: auto-focus and select input content on mount
const vAutoFocus = {
  mounted(el) {
    nextTick(() => {
      el.focus()
      el.select()
    })
  }
}

const SQLEditor = defineAsyncComponent(() => import('@/components/SQLEditor.vue'))
const DatabaseOverview = defineAsyncComponent(() => import('@/components/DatabaseOverview.vue'))
const ImportExport = defineAsyncComponent(() => import('@/components/ImportExport.vue'))
const SQLHistoryPanel = defineAsyncComponent(() => import('@/components/SQLHistoryPanel.vue'))
const TransferCenter = defineAsyncComponent(() => import('@/components/TransferCenter.vue'))

import { useEditMode } from '@/composables/useEditMode'
import { useQueryEditor } from '@/composables/useQueryEditor'

const SQL_HISTORY_KEY = 'sql-history'
const MAX_SQL_HISTORY = 50

const { t } = useI18n()
const store = useDatabaseStore()

// ── Composables ──
// deps is mutable: we assign ensureDataTabSchema and refreshTableData later
const editDeps = {}
const {
  startEdit, commitEdit, cancelEdit, getAdjacentCell,
  handleEditKeydown, handleCellBlur,
  hasPendingEdits, revertEdits, getCellValue,
  isRowEdited, isEditing, isCellModified,
  toggleEditMode, isRowSelected, toggleRowSelect,
  isAllSelected, toggleSelectAll, hasSelectedRows, isRowDeleted,
  addNewRow, startNewRowEdit, handleNewRowKeydown, commitNewRowEdit,
  isNewRowSelected, toggleNewRowSelect,
  markDeleteSelected, confirmSaveEdits, executeAllEdits,
  getColumnHint
} = useEditMode(t, editDeps)

const {
  parseQueryTableName, detectQueryEditable, ensureQueryTabSchema,
  toggleQueryEditMode, isQueryEditing, isQueryCellModified,
  isQueryRowEdited, isQueryRowSelected, isQueryRowDeleted,
  getQueryCellValue, startQueryEdit, commitQueryEdit,
  handleQueryEditKeydown, handleQueryCellBlur,
  toggleQueryRowSelect, toggleQuerySelectAll, isQueryAllSelected,
  addQueryNewRow, startQueryNewRowEdit, handleQueryNewRowKeydown,
  commitQueryNewRowEdit, toggleQueryNewRowSelect,
  markDeleteQuerySelected, hasQueryPendingEdits, countQueryEdits,
  revertQueryEdits, confirmSaveQueryEdits, executeQueryEdits
} = useQueryEditor(t)

const schemaOptions = ref([])
const loadingSchemas = ref(false)
const isPostgreSQL = computed(() => store.connectionConfig?.dbType === 'postgres')

const tabEditorRefs = reactive({})
const workspaceStateByConn = reactive({})

const ieVisible = ref(false)
const ieMode = ref('export')
const ieTableName = ref('')
const ieDatabase = ref('')
const ieConnId = ref('')
const iePageRows = ref([])
const iePageColumns = ref([])

const historyVisible = ref(false)
const historyTargetTabId = ref(null)
const tabSwitcherVisible = ref(false)
const tabSearch = ref('')
const workAreaRef = ref(null)

function buildWorkspaceSnapshot() {
  const byConn = {}
  let active = null

  Object.entries(workspaceStateByConn).forEach(([connId, workspace]) => {
    const summary = {
      overviews: [],
      objects: [],
      queries: []
    }

    for (const tab of workspace.tabs || []) {
      if (tab.type === 'overview' && tab.database) {
        summary.overviews.push(tab.database)
      }

      if ((tab.type === 'data' || tab.type === 'structure') && tab.database && tab.table) {
        summary.objects.push(`${tab.database}::${tab.table}`)
      }

      if (tab.type === 'query' && tab.database) {
        summary.queries.push(tab.database)
      }
    }

    byConn[connId] = summary

    if (connId === store.currentConnId) {
      const activeTab = (workspace.tabs || []).find((tab) => tab.id === workspace.activeTabId)
      if (activeTab) {
        active = {
          connId,
          type: activeTab.type,
          database: activeTab.database || '',
          table: activeTab.table || ''
        }
      }
    }
  })

  return { byConn, active }
}

function publishWorkspaceSnapshot() {
  window.dispatchEvent(new CustomEvent('workspace-tabs-updated', {
    detail: buildWorkspaceSnapshot()
  }))
}

function ensureWorkspace(connId) {
  if (!connId) return null

  if (!workspaceStateByConn[connId]) {
    workspaceStateByConn[connId] = reactive({
      tabIdCounter: 1,
      tabs: [],
      activeTabId: null
    })
  }

  return workspaceStateByConn[connId]
}

const currentWorkspace = computed(() => (
  store.currentConnId ? ensureWorkspace(store.currentConnId) : null
))

const workspaceTabs = computed(() => currentWorkspace.value?.tabs || [])
const activeTabId = computed({
  get: () => currentWorkspace.value?.activeTabId || null,
  set: (value) => {
    if (currentWorkspace.value) {
      currentWorkspace.value.activeTabId = value
    }
  }
})

const filteredWorkspaceTabs = computed(() => {
  const keyword = tabSearch.value.trim().toLowerCase()
  if (!keyword) return workspaceTabs.value

  return workspaceTabs.value.filter((tab) => (
    `${getTabPlainTitle(tab)} ${getTabPlainSubtitle(tab)} ${tab.type}`.toLowerCase().includes(keyword)
  ))
})

watch(() => store.currentConnId, (connId) => {
  if (connId) ensureWorkspace(connId)
}, { immediate: true })

watch(tabSwitcherVisible, (show) => {
  if (!show) {
    tabSearch.value = ''
  }
})

watch(
  () => [activeTabId.value, workspaceTabs.value.length, store.currentConnId],
  async () => {
    await nextTick()
    scrollActiveTabIntoView()
  }
)

watch(workspaceStateByConn, () => {
  publishWorkspaceSnapshot()
}, { deep: true })

watch(() => store.currentConnId, () => {
  publishWorkspaceSnapshot()
})

watch(
  () => store.connections.map((item) => item.connId),
  (connIds) => {
    Object.keys(workspaceStateByConn).forEach((connId) => {
      if (!connIds.includes(connId)) {
        delete workspaceStateByConn[connId]
      }
    })
    publishWorkspaceSnapshot()
  }
)

const schemaColumns = computed(() => [
  { title: t('schema.colName'), key: 'name' },
  { title: t('schema.colType'), key: 'type' },
  { title: t('schema.colNullable'), key: 'nullable', render: (row) => row.nullable ? t('schema.colNullableYes') : t('schema.colNullableNo') },
  { title: t('schema.colDefault'), key: 'defaultValue' },
  { title: t('schema.colPrimaryKey'), key: 'isPrimaryKey', render: (row) => row.isPrimaryKey ? '✓' : '' },
  { title: t('schema.colComment'), key: 'comment' }
])

const databaseOptions = computed(() => (
  (store.databases || []).map((db) => ({
    label: db.name,
    value: db.name
  }))
))

function nextTabId(connId) {
  const workspace = ensureWorkspace(connId)
  const id = `tab-${connId}-${workspace.tabIdCounter}`
  workspace.tabIdCounter += 1
  return id
}

function getWorkspace(connId = store.currentConnId) {
  return connId ? ensureWorkspace(connId) : null
}

function getSqlHistory() {
  try {
    return JSON.parse(localStorage.getItem(SQL_HISTORY_KEY) || '[]')
  } catch {
    return []
  }
}

function saveSqlHistory(sql, database, connId) {
  const normalizedSql = (sql || '').trim()
  if (!normalizedSql) return

  const history = getSqlHistory().filter((item) => !(
    item.sql === normalizedSql &&
    item.database === database &&
    item.connId === connId
  ))

  history.unshift({
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    sql: normalizedSql,
    summary: normalizedSql.replace(/\s+/g, ' ').slice(0, 120),
    database: database || '',
    connId: connId || '',
    timestamp: Date.now()
  })

  localStorage.setItem(SQL_HISTORY_KEY, JSON.stringify(history.slice(0, MAX_SQL_HISTORY)))
}

function openHistoryPanel(tabId) {
  historyTargetTabId.value = tabId
  historyVisible.value = true
}

function handleHistorySelect(sql) {
  const editor = tabEditorRefs[historyTargetTabId.value]
  const currentTab = workspaceTabs.value.find((tab) => tab.id === historyTargetTabId.value)

  if (editor?.insertText) {
    editor.insertText(sql)
  } else if (currentTab) {
    currentTab.sql = currentTab.sql ? `${currentTab.sql}\n${sql}` : sql
  }
}

function renderDbLabel(option) {
  return h(NTooltip, { trigger: 'hover', placement: 'right' }, {
    trigger: () => h('span', {
      style: 'display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;'
    }, option.label),
    default: () => option.label
  })
}

function getTabPlainTitle(tab) {
  if (tab.type === 'query') return tab.name || t('query.tabName', { n: '' })
  if (tab.type === 'data') return tab.table || tab.name || ''
  if (tab.type === 'overview') return tab.database || tab.name || ''
  if (tab.type === 'structure') return tab.table || tab.name || ''
  if (tab.type === 'transfer') return tab.name || t('transfer.tabName')
  return tab.name || ''
}

function getTabPlainSubtitle(tab) {
  if (tab.type === 'query') {
    return tab.database ? `${t('query.selectDb')} · ${tab.database}` : t('query.noDbHint')
  }
  if (tab.type === 'overview') {
    return t('overview.tabShort')
  }
  if (tab.database && tab.table) {
    return `${tab.database} / ${tab.table}`
  }
  if (tab.database) {
    return tab.database
  }
  return tab.type
}

function selectWorkspaceTab(tabId) {
  activeTabId.value = tabId
  tabSwitcherVisible.value = false
}

function scrollActiveTabIntoView() {
  const root = workAreaRef.value
  if (!root) return

  const activeTabEl = root.querySelector('.workspace-tabs .n-tabs-tab.n-tabs-tab--active')
  if (!activeTabEl || typeof activeTabEl.scrollIntoView !== 'function') return

  activeTabEl.scrollIntoView({
    block: 'nearest',
    inline: 'nearest',
    behavior: 'smooth'
  })
}

function getTabLabel(tab) {
  const iconStyle = 'width: 13px; height: 13px; margin-right: 6px; opacity: 0.7; flex-shrink: 0;'
  const dimStyle = 'opacity: 0.45; font-size: 12px;'

  if (tab.type === 'query') {
    return h('span', { style: 'display:inline-flex;align-items:center;' }, [
      h('svg', { width: 13, height: 13, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round', style: iconStyle }, [
        h('polyline', { points: '4 17 10 11 4 5' }),
        h('line', { x1: '12', y1: '19', x2: '20', y2: '19' })
      ]),
      tab.name
    ])
  }

  if (tab.type === 'data') {
    return h('span', { style: 'display:inline-flex;align-items:center;' }, [
      h('svg', { width: 13, height: 13, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round', style: iconStyle }, [
        h('rect', { x: '3', y: '3', width: '18', height: '18', rx: '2' }),
        h('path', { d: 'M3 9h18M3 15h18M9 3v18M15 3v18' })
      ]),
      tab.table,
      h('span', { style: dimStyle }, `@${tab.database}`)
    ])
  }

  if (tab.type === 'overview') {
    return h('span', { style: 'display:inline-flex;align-items:center;' }, [
      h('svg', { width: 13, height: 13, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round', style: iconStyle }, [
        h('ellipse', { cx: '12', cy: '5', rx: '8', ry: '3' }),
        h('path', { d: 'M4 5v6c0 1.66 3.58 3 8 3s8-1.34 8-3V5' }),
        h('path', { d: 'M4 11v6c0 1.66 3.58 3 8 3s8-1.34 8-3v-6' })
      ]),
      tab.database,
      h('span', { style: dimStyle }, t('overview.tabShort'))
    ])
  }

  if (tab.type === 'structure') {
    return h('span', { style: 'display:inline-flex;align-items:center;' }, [
      h('svg', { width: 13, height: 13, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round', style: iconStyle }, [
        h('path', { d: 'M12 2L2 7l10 5 10-5-10-5z' }),
        h('path', { d: 'M2 17l10 5 10-5' }),
        h('path', { d: 'M2 12l10 5 10-5' })
      ]),
      tab.table,
      h('span', { style: dimStyle }, `@${tab.database}`)
    ])
  }

  if (tab.type === 'transfer') {
    return h('span', { style: 'display:inline-flex;align-items:center;' }, [
      h('svg', { width: 13, height: 13, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round', style: iconStyle }, [
        h('path', { d: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4' }),
        h('polyline', { points: '17 8 12 3 7 8' }),
        h('line', { x1: '12', y1: '3', x2: '12', y2: '15' })
      ]),
      tab.name
    ])
  }

  return tab.name
}

function buildCompletionSchema(database) {
  const completionSchema = {}

  if (database === store.selectedDatabase) {
    for (const table of store.tables || []) {
      completionSchema[table.name] = {}
    }
  }

  if (
    database === store.selectedDatabase &&
    store.selectedTable &&
    store.tableSchema?.columns?.length
  ) {
    completionSchema[store.selectedTable] = store.tableSchema.columns.map((column) => column.name)
  }

  for (const tab of workspaceTabs.value) {
    if (tab.database !== database || !tab.table) continue

    const schema = Array.isArray(tab.schema) ? tab.schema : tab.schema?.columns
    if (schema?.length) {
      completionSchema[tab.table] = schema.map((column) => column.name)
    }

    if (tab.schemaInfo?.columns?.length) {
      completionSchema[tab.table] = tab.schemaInfo.columns.map((column) => column.name)
    }
  }

  return completionSchema
}

function getCompletionSchema(database) {
  return buildCompletionSchema(database)
}

async function handleDatabaseChange(tab) {
  tab.schema = null
  schemaOptions.value = []

  if (!tab.database || !tab.connId) return

  try {
    await store.selectDatabase(tab.database, tab.connId)
  } catch (error) {
    gmssh.error(t('query.executeFailed', { msg: error.message }))
    return
  }

  if (!isPostgreSQL.value) return

  loadingSchemas.value = true
  try {
    const schemas = await api.listSchemas(tab.connId, tab.database)
    schemaOptions.value = (schemas || []).map((schema) => ({ label: schema, value: schema }))

    if (schemaOptions.value.find((schema) => schema.value === 'public')) {
      tab.schema = 'public'
    } else if (schemaOptions.value.length > 0) {
      tab.schema = schemaOptions.value[0].value
    }
  } catch (error) {
    gmssh.error(t('query.executeFailed', { msg: error.message }))
  } finally {
    loadingSchemas.value = false
  }
}

function createQueryTab(connId, database = null, sql = '') {
  const workspace = getWorkspace(connId)
  const connState = store.getConnectionState(connId)
  return reactive({
    id: nextTabId(connId),
    connId,
    type: 'query',
    name: t('query.tabName', { n: workspace.tabIdCounter - 1 }),
    database: database || connState?.selectedDatabase || null,
    schema: null,
    sql,
    result: null,
    executing: false,
    editorHeight: 55,
    // Query result editing state
    queryEditMode: false,
    queryEditTable: null,
    queryPendingEdits: {},
    queryEditingCell: null,
    queryNewRows: [],
    queryEditingNewCell: null,
    querySelectedRows: new Set(),
    querySelectedNewRows: new Set(),
    queryDeletedRows: new Set(),
    queryPrimaryKeys: [],
    queryCanEdit: null
  })
}

function createOverviewTab(connId, database) {
  return reactive({
    id: nextTabId(connId),
    connId,
    type: 'overview',
    database,
    overviewCache: {
      loaded: false,
      loading: false,
      objects: []
    }
  })
}

function ensureOverviewCache(tab) {
  if (!tab || tab.type !== 'overview') return null

  if (!tab.overviewCache) {
    tab.overviewCache = reactive({
      loaded: false,
      loading: false,
      objects: []
    })
    return tab.overviewCache
  }

  if (!Array.isArray(tab.overviewCache.objects)) {
    tab.overviewCache.objects = []
  }
  if (typeof tab.overviewCache.loaded !== 'boolean') {
    tab.overviewCache.loaded = false
  }
  if (typeof tab.overviewCache.loading !== 'boolean') {
    tab.overviewCache.loading = false
  }

  return tab.overviewCache
}

function createDataTab(connId, database, table) {
  return reactive({
    id: nextTabId(connId),
    connId,
    type: 'data',
    database,
    table,
    tableData: null,
    loading: false,
    page: 1,
    pageSize: 500,
    sortCol: '',
    sortDir: '',
    filterWhere: '',
    filterInput: '',
    showFilter: false,
    pendingEdits: {},
    editingCell: null,
    schemaInfo: null,
    metaInfo: null,
    primaryKeys: [],
    canEdit: null,
    ddlText: '',
    ddlLoading: false,
    // Edit mode state
    editMode: false,
    activeRowIndex: null,
    formViewMode: false,
    selectedRows: new Set(),
    deletedRows: new Set(),
    newRows: [],
    editingNewCell: null,
    selectedNewRows: new Set()
  })
}

function createStructureTab(connId, database, table) {
  return reactive({
    id: nextTabId(connId),
    connId,
    type: 'structure',
    database,
    table,
    schema: null,
    loading: false,
    // Inspector sidebar state
    showInspector: false,
    metaInfo: null,
    ddlText: '',
    ddlLoading: false
  })
}

function createTransferTab(connId, database = '', action = '', autoLaunch = false) {
  return reactive({
    id: nextTabId(connId),
    connId,
    type: 'transfer',
    database,
    action,
    autoLaunch,
    launchKey: Date.now(),
    name: t('transfer.tabName')
  })
}

function closeTab(tabId, connId = store.currentConnId) {
  const workspace = getWorkspace(connId)
  if (!workspace) return

  const index = workspace.tabs.findIndex((tab) => tab.id === tabId)
  if (index === -1) return

  workspace.tabs.splice(index, 1)
  delete tabEditorRefs[tabId]

  if (workspace.activeTabId === tabId) {
    workspace.activeTabId = workspace.tabs[Math.max(0, index - 1)]?.id || workspace.tabs[0]?.id || null
  }
}

function setTabEditorRef(tabId, el) {
  if (el) {
    tabEditorRefs[tabId] = el
  } else {
    delete tabEditorRefs[tabId]
  }
}

function openExportModal(tableName, database, context = {}) {
  if (!tableName) return
  ieMode.value = 'export'
  ieTableName.value = tableName
  ieDatabase.value = database || store.selectedDatabase
  ieConnId.value = context.connId || store.currentConnId || store.activeConnection
  iePageRows.value = context.rows || []
  iePageColumns.value = context.columns || []
  ieVisible.value = true
}

function openImportModal(tableName, database, context = {}) {
  if (!tableName) return
  ieMode.value = 'import'
  ieTableName.value = tableName
  ieDatabase.value = database || store.selectedDatabase
  ieConnId.value = context.connId || store.currentConnId || store.activeConnection
  iePageRows.value = []
  iePageColumns.value = []
  ieVisible.value = true
}

async function ensureDataTabSchema(tab) {
  if (tab.schemaInfo) {
    return tab.schemaInfo
  }

  const schema = await api.getTableSchema(tab.connId, tab.database, tab.table)
  tab.schemaInfo = schema
  tab.primaryKeys = schema?.primaryKey || []
  tab.canEdit = tab.primaryKeys.length > 0
  if (tab.metaInfo) {
    tab.metaInfo = {
      ...tab.metaInfo,
      columnCount: Array.isArray(schema?.columns) ? schema.columns.length : tab.metaInfo.columnCount
    }
  }
  return schema
}

async function openTableDataTab(database, table, connId = store.currentConnId) {
  if (!connId) return

  const workspace = getWorkspace(connId)
  const existing = workspace.tabs.find((tab) => (
    tab.type === 'data' &&
    tab.database === database &&
    tab.table === table
  ))

  if (existing) {
    workspace.activeTabId = existing.id
    return existing
  }

  const newTab = createDataTab(connId, database, table)
  workspace.tabs.push(newTab)
  workspace.activeTabId = newTab.id

  await Promise.all([
    refreshTableData(newTab),
    ensureDataTabSchema(newTab).catch(() => null),
    loadTableMetaInfo(newTab).catch(() => null),
    loadTableDDL(newTab).catch(() => null)
  ])

  return newTab
}

async function openStructureTab(database, table, connId = store.currentConnId) {
  if (!connId) return

  const workspace = getWorkspace(connId)
  const existing = workspace.tabs.find((tab) => (
    tab.type === 'structure' &&
    tab.database === database &&
    tab.table === table
  ))

  if (existing) {
    workspace.activeTabId = existing.id
    return existing
  }

  const newTab = createStructureTab(connId, database, table)
  workspace.tabs.push(newTab)
  workspace.activeTabId = newTab.id
  await refreshTableStructure(newTab)
  return newTab
}

async function openOverviewTab(database, connId = store.currentConnId) {
  if (!connId || !database) return

  const workspace = getWorkspace(connId)
  const existing = workspace.tabs.find((tab) => (
    tab.type === 'overview' &&
    tab.database === database
  ))

  if (existing) {
    ensureOverviewCache(existing)
    workspace.activeTabId = existing.id
    return
  }

  const newTab = createOverviewTab(connId, database)
  workspace.tabs.push(newTab)
  workspace.activeTabId = newTab.id
  refreshOverviewTab(newTab).catch(() => null)
}

async function refreshOverviewTab(tab) {
  if (!tab?.connId || !tab?.database) return

  const cache = ensureOverviewCache(tab)
  if (!cache) return

  cache.loading = true
  try {
    const tableList = await api.listTables(tab.connId, tab.database)
    cache.objects = Array.isArray(tableList) ? tableList : []
    cache.loaded = true
  } catch (error) {
    gmssh.error(error.message || String(error))
  } finally {
    cache.loading = false
  }
}

function openQueryTabWithContext(database, table = null, connId = store.currentConnId, sql = '') {
  if (!connId) return

  const initialSql = sql || (table ? `SELECT * FROM \`${table}\`` : '')
  const workspace = getWorkspace(connId)
  const newTab = createQueryTab(connId, database, initialSql)
  workspace.tabs.push(newTab)
  workspace.activeTabId = newTab.id
}

function openTransferTab(connId = store.currentConnId, database = '', action = '', autoLaunch = false) {
  if (!connId) return null

  const workspace = getWorkspace(connId)
  const existing = workspace.tabs.find((tab) => tab.type === 'transfer')

  if (existing) {
    existing.connId = connId
    existing.database = database || ''
    existing.action = action || ''
    existing.autoLaunch = !!autoLaunch
    if (autoLaunch) {
      existing.launchKey = Date.now()
    }
    workspace.activeTabId = existing.id
    return existing
  }

  const newTab = createTransferTab(connId, database, action, autoLaunch)
  workspace.tabs.push(newTab)
  workspace.activeTabId = newTab.id
  return newTab
}

function isNumericCol(tab, col) {
  const rows = tab.result?.rows
  if (!rows?.length) return false

  for (const row of rows) {
    const value = row[col]
    if (value == null) continue
    return typeof value === 'number' || (typeof value === 'string' && /^-?\d+(\.\d+)?$/.test(value.trim()))
  }

  return false
}

function isNumericCol_data(tab, col) {
  const rows = tab.tableData?.rows
  if (!rows?.length) return false

  for (const row of rows) {
    const value = row[col]
    if (value == null) continue
    return typeof value === 'number' || (typeof value === 'string' && /^-?\d+(\.\d+)?$/.test(value.trim()))
  }

  return false
}

function startSplitterDrag(event, tab) {
  event.preventDefault()
  const layoutEl = event.target.parentElement
  const startY = event.clientY
  const startHeight = tab.editorHeight || 55
  const totalHeight = layoutEl.clientHeight

  const onMove = (moveEvent) => {
    const delta = moveEvent.clientY - startY
    const percentage = startHeight + (delta / totalHeight * 100)
    tab.editorHeight = Math.min(85, Math.max(20, percentage))
  }

  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.body.style.cursor = 'row-resize'
  document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

function stopTabExecution(tab) {
  tab.executing = false
  gmssh.info(t('query.stop'))
}

function formatDuration(ms) {
  if (ms == null) return '–'
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatInspectorValue(value) {
  if (value == null || value === '') return '—'
  return String(value)
}

function formatSize(size) {
  if (!size) return '—'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

async function copyToClipboard(text, successMessage) {
  if (!text) return

  const copied = await gmssh.copyToClipboard(text)
  if (copied) {
    gmssh.success(successMessage)
    return
  }

  gmssh.warning(t('designer.copyUnavailable'))
}

function toggleInspector(tab) {
  tab.showInspector = !tab.showInspector

  if (tab.showInspector) {
    if (!tab.metaInfo) {
      loadTableMetaInfo(tab).catch(() => null)
    }
    if (!tab.ddlText) {
      loadTableDDL(tab).catch(() => null)
    }
  }
}

async function loadTableMetaInfo(tab) {
  const objects = await api.listTables(tab.connId, tab.database)
  const object = (objects || []).find((item) => item.name === tab.table)

  tab.metaInfo = object ? {
    rowCount: object.rowCount,
    size: object.size,
    engine: object.engine,
    createdAt: object.createdAt,
    updatedAt: object.updatedAt,
    comment: object.comment,
    columnCount: Array.isArray(tab.schemaInfo?.columns) ? tab.schemaInfo.columns.length : undefined
  } : {
    columnCount: Array.isArray(tab.schemaInfo?.columns) ? tab.schemaInfo.columns.length : undefined
  }

  return tab.metaInfo
}

async function loadTableDDL(tab, force = false) {
  if (!force && tab.ddlText) return tab.ddlText

  tab.ddlLoading = true
  try {
    tab.ddlText = await api.getTableDDL(tab.connId, tab.database, tab.table)
    return tab.ddlText
  } catch (error) {
    gmssh.error(error.message)
    return ''
  } finally {
    tab.ddlLoading = false
  }
}

async function copyInspectorDDL(tab) {
  if (!tab.ddlText) {
    await loadTableDDL(tab)
  }

  if (tab.ddlText) {
    await copyToClipboard(tab.ddlText, t('overview.ddlCopied'))
  }
}

async function executeTabSQL(tab) {
  if (!tab.sql.trim()) {
    gmssh.warning(t('query.sqlRequired'))
    return
  }

  if (!tab.database) {
    gmssh.warning(t('query.dbRequired'))
    return
  }

  tab.executing = true
  tab.result = null

  try {
    const editorRef = tabEditorRefs[tab.id]
    let sql = ''

    if (editorRef?.getExecutionSQL) {
      sql = editorRef.getExecutionSQL()
    } else if (editorRef?.getSelectedText) {
      sql = editorRef.getSelectedText()
    }

    sql = (sql || tab.sql || '').trim()

    const isSingleSelect = sql.toLowerCase().startsWith('select') &&
      !sql.slice(0, -1).includes(';') &&
      !/\blimit\s+\d+/i.test(sql)

    if (isSingleSelect) {
      sql = sql.replace(/;?\s*$/, '') + ' LIMIT 100'
    }

    const result = await store.executeSQL(sql, tab.database, tab.connId)

    if (result.success) {
      const rows = []

      if (result.data?.rows && result.data?.columns) {
        for (const row of result.data.rows) {
          const rowObject = {}
          result.data.columns.forEach((column, index) => {
            rowObject[column] = row[index]
          })
          rows.push(rowObject)
        }
      }

      tab.result = {
        success: true,
        columns: result.data?.columns || [],
        rows,
        rowCount: rows.length,
        executionTime: result.executionTime ?? 0
      }

      // Reset query editing state and detect if editable
      tab.queryCanEdit = null
      tab.queryEditMode = false
      tab.queryPendingEdits = {}
      tab.queryNewRows = []
      tab.queryDeletedRows = new Set()
      tab.querySelectedRows = new Set()
      tab.queryPrimaryKeys = []
      detectQueryEditable(tab)

      saveSqlHistory(sql, tab.database, tab.connId)
      gmssh.success(t('query.executeSuccess'))
      return
    }

    tab.result = {
      success: false,
      error: result.error || t('query.executeFailed', { msg: '' })
    }
    gmssh.error(tab.result.error)
  } catch (error) {
    tab.result = {
      success: false,
      error: error.message
    }
    gmssh.error(t('query.executeFailed', { msg: error.message }))
  } finally {
    tab.executing = false
  }
}

// ── Form View Edit Helpers ──
function getFormCellValue(tab, rowIdx, col) {
  const val = getCellValue(tab, tab.tableData.rows[rowIdx], rowIdx, col)
  return val == null ? '' : String(val)
}

function commitFormEdit(event, tab, rowIdx, col) {
  const newValue = event.target.value
  const originalValue = tab.tableData.rows[rowIdx][col]
  const originalStr = originalValue == null ? '' : String(originalValue)

  if (newValue === originalStr) {
    // No change — remove from pendingEdits if it was there
    if (tab.pendingEdits[rowIdx]) {
      delete tab.pendingEdits[rowIdx][col]
      if (Object.keys(tab.pendingEdits[rowIdx]).length === 0) {
        delete tab.pendingEdits[rowIdx]
      }
      tab.pendingEdits = { ...tab.pendingEdits }
    }
    return
  }

  if (!tab.pendingEdits[rowIdx]) {
    tab.pendingEdits[rowIdx] = {}
  }
  tab.pendingEdits[rowIdx][col] = newValue === '' ? null : newValue
  tab.pendingEdits = { ...tab.pendingEdits }
}

function autoResizeTextarea(event) {
  const el = event.target
  el.style.height = 'auto'
  el.style.height = el.scrollHeight + 'px'
}

async function handleAddNewRow(tab) {
  await addNewRow(tab)
  // Auto-scroll to bottom after adding new row
  await nextTick()
  const tabEl = workAreaRef.value?.querySelector?.('.premium-table-wrap')
  if (tabEl) {
    tabEl.scrollTop = tabEl.scrollHeight
  }
}

async function refreshTableData(tab) {
  tab.loading = true

  try {
    const data = await api.getTableData(
      tab.connId,
      tab.database,
      tab.table,
      tab.page,
      tab.pageSize,
      tab.sortCol || '',
      tab.sortDir || ''
    )

    const rows = (data.rows || []).map((row) => {
      const rowObject = {}
      data.columns.forEach((column, index) => {
        rowObject[column] = row[index]
      })
      return rowObject
    })

    tab.tableData = { ...data, rows }
    tab.pendingEdits = {}
    tab.editingCell = null
    if (tab.metaInfo) {
      tab.metaInfo = {
        ...tab.metaInfo,
        rowCount: data.total,
        columnCount: data.columns?.length || tab.metaInfo.columnCount
      }
    }
  } catch (error) {
    gmssh.error(t('dataTab.loadFailed', { msg: error.message }))
  } finally {
    tab.loading = false
  }
}

function jumpToPage(tab, pageInput) {
  const totalPages = Math.max(1, Math.ceil((tab.tableData?.total || 0) / tab.pageSize))
  let p = parseInt(pageInput, 10)
  if (isNaN(p) || p < 1) p = 1
  if (p > totalPages) p = totalPages
  if (p === tab.page) return
  tab.page = p
  refreshTableData(tab)
}

function toggleSort(tab, col) {
  if (tab.sortCol === col) {
    tab.sortDir = tab.sortDir === 'asc' ? 'desc' : tab.sortDir === 'desc' ? '' : 'asc'
    if (!tab.sortDir) tab.sortCol = ''
  } else {
    tab.sortCol = col
    tab.sortDir = 'asc'
  }

  tab.page = 1
  refreshTableData(tab)
}

function applyFilter(tab) {
  tab.filterWhere = tab.filterInput.trim()
  tab.page = 1
  refreshTableData(tab)
}

function clearFilter(tab) {
  tab.filterWhere = ''
  tab.filterInput = ''
  tab.page = 1
  refreshTableData(tab)
}

// Late-bind dependencies for useEditMode composable
editDeps.ensureDataTabSchema = ensureDataTabSchema
editDeps.refreshTableData = refreshTableData

async function refreshTableStructure(tab) {
  tab.loading = true
  try {
    const schema = await store.getTableSchemaFor(tab.table, tab.database, tab.connId)
    tab.schema = Array.isArray(schema) ? schema : (schema?.columns || [])

    // Load metaInfo for inspector sidebar (if structure tab has it)
    if ('metaInfo' in tab) {
      try {
        const tables = await api.listTables(tab.connId, tab.database)
        const meta = (tables || []).find(t2 => t2.name === tab.table)
        if (meta) {
          tab.metaInfo = {
            rowCount: meta.rows,
            columnCount: tab.schema?.length,
            engine: meta.engine,
            size: meta.size,
            createdAt: meta.createdAt,
            updatedAt: meta.updatedAt,
            comment: meta.comment
          }
        }
      } catch (_) { /* ignore meta errors */ }

      // Auto-load DDL
      if (!tab.ddlText) {
        loadTableDDL(tab, false)
      }
    }
  } catch (error) {
    gmssh.error(t('dataTab.structureFailed', { msg: error.message }))
  } finally {
    tab.loading = false
  }
}

function findDataTab(database, table, connId = store.currentConnId) {
  const workspace = getWorkspace(connId)
  if (!workspace) return null
  return workspace.tabs.find((tab) => (
    tab.type === 'data' &&
    tab.database === database &&
    tab.table === table
  )) || null
}

function handleOpenTableData(event) {
  const { database, table, connId } = event.detail
  openTableDataTab(database, table, connId || store.currentConnId)
}

function handleOpenStructure(event) {
  const { database, table, connId } = event.detail
  openStructureTab(database, table, connId || store.currentConnId)
}

function handleOpenQuery(event) {
  const { database, table, connId, sql } = event.detail || {}
  openQueryTabWithContext(database, table, connId || store.currentConnId, sql || '')
}

async function handleOpenObjectProperties(event) {
  const { database, table, connId } = event.detail || {}
  const targetConnId = connId || store.currentConnId
  if (!database || !table || !targetConnId) return

  const existing = findDataTab(database, table, targetConnId)
  const tab = existing || await openTableDataTab(database, table, targetConnId)
  if (tab) {
    tab.showInspector = true
    if (!tab.metaInfo) {
      loadTableMetaInfo(tab).catch(() => null)
    }
    if (!tab.ddlText) {
      loadTableDDL(tab).catch(() => null)
    }
  }
}

async function handleTableRenamed(event) {
  const { database, oldName, newName, connId } = event.detail || {}
  const targetConnId = connId || store.currentConnId
  const workspace = getWorkspace(targetConnId)
  if (!workspace || !database || !oldName || !newName) return

  for (const tab of workspace.tabs) {
    if (tab.database !== database || tab.table !== oldName) continue

    tab.table = newName

    if (tab.type === 'data') {
      tab.ddlText = ''
      tab.metaInfo = null
      await Promise.all([
        refreshTableData(tab).catch(() => null),
        ensureDataTabSchema(tab).catch(() => null),
        loadTableMetaInfo(tab).catch(() => null),
        loadTableDDL(tab).catch(() => null)
      ])
    } else if (tab.type === 'structure') {
      await refreshTableStructure(tab)
    }
  }
}

function handleSwitchToSql() {
  openQueryTabWithContext(store.selectedDatabase, store.selectedTable)
}

async function handleOpenDatabaseOverview(event) {
  const { database, connId } = event.detail || {}
  if (!database) return

  if (connId) {
    try {
      await store.selectDatabase(database, connId)
    } catch {
      // DatabaseOverview refreshes itself, so a preselect failure should not block tab creation.
    }
  }

  openOverviewTab(database, connId || store.currentConnId)
}

function handleOpenExport(event) {
  const { database, table, connId } = event.detail || {}
  const targetConnId = connId || store.currentConnId
  const dataTab = findDataTab(database, table, targetConnId)
  openExportModal(table, database, {
    connId: targetConnId,
    rows: dataTab?.tableData?.rows || [],
    columns: dataTab?.tableData?.columns || []
  })
}

function handleOpenImport(event) {
  const { database, table, connId } = event.detail || {}
  openImportModal(table, database, { connId: connId || store.currentConnId })
}

function handleOpenTransferCenter(event) {
  const {
    connId,
    database = '',
    action = '',
    autoLaunch = false
  } = event.detail || {}

  openTransferTab(connId || store.currentConnId, database, action, autoLaunch)
}

async function handleDatabaseDumpFinished(event) {
  const { connId, database } = event.detail || {}
  const targetConnId = connId || store.currentConnId
  if (!targetConnId || !database) return

  const workspace = getWorkspace(targetConnId)
  if (!workspace) return

  const refreshJobs = []

  for (const tab of workspace.tabs) {
    if (tab.connId !== targetConnId || tab.database !== database) continue

    if (tab.type === 'overview') {
      refreshJobs.push(refreshOverviewTab(tab).catch(() => null))
      continue
    }

    if (tab.type === 'structure') {
      refreshJobs.push(refreshTableStructure(tab).catch(() => null))
      continue
    }

    if (tab.type === 'data') {
      tab.schemaInfo = null
      tab.primaryKeys = []
      tab.canEdit = null
      tab.metaInfo = null
      tab.ddlText = ''
      refreshJobs.push(Promise.all([
        refreshTableData(tab).catch(() => null),
        ensureDataTabSchema(tab).catch(() => null),
        loadTableMetaInfo(tab).catch(() => null),
        loadTableDDL(tab).catch(() => null)
      ]))
    }
  }

  await Promise.all(refreshJobs)
}

async function handleRefreshTableData(event) {
  const { database, table, connId } = event.detail || {}
  const targetConnId = connId || store.currentConnId
  if (!database || !table || !targetConnId) return

  const dataTab = findDataTab(database, table, targetConnId)
  if (!dataTab) return

  await Promise.all([
    refreshTableData(dataTab).catch(() => null),
    loadTableMetaInfo(dataTab).catch(() => null)
  ])
}

onMounted(() => {
  publishWorkspaceSnapshot()
  window.addEventListener('open-database-overview', handleOpenDatabaseOverview)
  window.addEventListener('open-table-data', handleOpenTableData)
  window.addEventListener('open-structure', handleOpenStructure)
  window.addEventListener('open-query', handleOpenQuery)
  window.addEventListener('open-object-properties', handleOpenObjectProperties)
  window.addEventListener('table-renamed', handleTableRenamed)
  window.addEventListener('switch-to-sql-tab', handleSwitchToSql)
  window.addEventListener('open-export', handleOpenExport)
  window.addEventListener('open-import', handleOpenImport)
  window.addEventListener('open-transfer-center', handleOpenTransferCenter)
  window.addEventListener('database-dump-finished', handleDatabaseDumpFinished)
  window.addEventListener('refresh-table-data', handleRefreshTableData)
})

onUnmounted(() => {
  window.dispatchEvent(new CustomEvent('workspace-tabs-updated', {
    detail: { byConn: {}, active: null }
  }))
  window.removeEventListener('open-database-overview', handleOpenDatabaseOverview)
  window.removeEventListener('open-table-data', handleOpenTableData)
  window.removeEventListener('open-structure', handleOpenStructure)
  window.removeEventListener('open-query', handleOpenQuery)
  window.removeEventListener('open-object-properties', handleOpenObjectProperties)
  window.removeEventListener('table-renamed', handleTableRenamed)
  window.removeEventListener('switch-to-sql-tab', handleSwitchToSql)
  window.removeEventListener('open-export', handleOpenExport)
  window.removeEventListener('open-import', handleOpenImport)
  window.removeEventListener('open-transfer-center', handleOpenTransferCenter)
  window.removeEventListener('database-dump-finished', handleDatabaseDumpFinished)
  window.removeEventListener('refresh-table-data', handleRefreshTableData)
})
</script>

<style scoped>
.work-area {
  height: 100%;
  padding: 0;
  background: transparent;
}

.workspace-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.workspace-tabs :deep(.n-tabs-nav) {
  display: flex;
  align-items: flex-end;
  gap: 0;
  margin-bottom: 0;
  padding: 0 6px;
  background: transparent !important;
  box-shadow: none !important;
  border-bottom: 1px solid #303640;
}

.workspace-tabs :deep(.n-tabs-nav-scroll-wrapper) {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  background: transparent !important;
}

.workspace-tabs :deep(.n-tabs-nav-scroll-content) {
  gap: 2px;
}

.workspace-tabs :deep(.n-tabs-nav__suffix) {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  padding-left: 8px;
}

/* ── Inactive Tab ── */
.workspace-tabs :deep(.n-tabs-tab) {
  border-radius: 6px 6px 0 0 !important;
  background: transparent !important;
  color: #A8ADB7 !important;
  border: 1px solid transparent !important;
  border-bottom: none !important;
  height: 36px !important;
  padding: 0 14px !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  transition: background 0.15s ease, color 0.15s ease, box-shadow 0.15s ease !important;
}

/* ── Active Tab: matches sidebar selected bg ── */
.workspace-tabs :deep(.n-tabs-tab.n-tabs-tab--active) {
  background: var(--ref-color-white-8, rgba(255, 255, 255, 0.08)) !important;
  color: #FFFFFF !important;
  border: 1px solid rgba(255, 255, 255, 0.06) !important;
  border-bottom: 1px solid transparent !important;
  box-shadow: inset 0 2px 0 var(--sys-color-border-focus) !important;
}

/* ── Close Button: default weakened ── */
.workspace-tabs :deep(.n-tabs-tab .n-base-close) {
  opacity: 0.4;
  border-radius: 4px;
  width: 18px;
  height: 18px;
  transition: opacity 0.15s ease, background 0.15s ease !important;
}

.workspace-tabs :deep(.n-tabs-tab .n-base-close:hover) {
  opacity: 1;
  background: #343B46;
}

.workspace-tabs :deep(.n-tabs-pane-wrapper) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: transparent !important;
}

.workspace-tabs :deep(.n-tab-pane) {
  height: 100%;
  padding: 8px 6px 0;
  background: transparent !important;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tab-switcher-trigger {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #303640;
  border-radius: 6px;
  background: #252A31;
  color: #A8ADB7;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.tab-switcher-trigger:hover {
  background: #343B46;
  color: #FFFFFF;
}

.tab-switcher-panel {
  width: 336px;
  max-height: min(70vh, 520px);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tab-switcher-search {
  display: block;
}

.tab-switcher-search :deep(.n-input) {
  --n-color: rgba(255, 255, 255, 0.05);
  --n-color-focus: rgba(255, 255, 255, 0.07);
  --n-border: 1px solid rgba(255, 255, 255, 0.06);
  --n-border-hover: 1px solid rgba(87, 114, 255, 0.25);
  --n-border-focus: 1px solid rgba(87, 114, 255, 0.38);
}

.tab-switcher-search :deep(.n-input .n-input__prefix) {
  color: var(--sys-color-text-tertiary);
}

.tab-switcher-section-title {
  color: var(--sys-color-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.tab-switcher-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow: auto;
  padding-right: 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.16) transparent;
}

.tab-switcher-list::-webkit-scrollbar {
  width: 8px;
}

.tab-switcher-list::-webkit-scrollbar-track {
  background: transparent;
}

.tab-switcher-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.16);
  border-radius: 999px;
}

.tab-switcher-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 10px 10px 12px;
  border: 1px solid transparent;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.03);
  cursor: pointer;
  text-align: left;
  transition: background 0.18s ease, border-color 0.18s ease, transform 0.18s ease;
  outline: none;
}

.tab-switcher-item:hover,
.tab-switcher-item:focus-visible {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.06);
}

.tab-switcher-item--active {
  background: rgba(87, 114, 255, 0.1);
  border-color: rgba(87, 114, 255, 0.2);
}

.tab-switcher-item__icon {
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--sys-color-text-secondary);
}

.tab-switcher-item--active .tab-switcher-item__icon {
  color: rgba(133, 154, 255, 0.95);
  background: rgba(87, 114, 255, 0.16);
}

.tab-switcher-item__content {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tab-switcher-item__title {
  color: var(--sys-color-text-primary);
  font-size: 13px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-switcher-item__meta {
  color: var(--sys-color-text-tertiary);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-switcher-item__close {
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--sys-color-text-tertiary);
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.tab-switcher-item__close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--sys-color-text-primary);
}

.tab-switcher-empty {
  padding: 16px 0 6px;
}

:deep(.tab-switcher-popover) {
  padding: 10px !important;
  border-radius: 8px !important;
  border: 1px solid rgba(255, 255, 255, 0.09) !important;
  background: rgba(40, 40, 40, 0.75) !important;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4) !important;
  backdrop-filter: blur(20px) !important;
  -webkit-backdrop-filter: blur(20px) !important;
}

.overview-tab-pane {
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.14) transparent;
}

.overview-tab-pane::-webkit-scrollbar {
  width: 5px;
}

.overview-tab-pane::-webkit-scrollbar-track {
  background: transparent;
}

.overview-tab-pane::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.14);
  border-radius: 999px;
}

.overview-tab-pane::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.24);
}

.field-label {
  color: var(--sys-color-text-secondary);
  font-size: var(--ref-font-size-sm);
}

/* Query Layout (Navicat-style) */
.query-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 0;
}

.query-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--ref-space-8) var(--ref-space-10);
  margin-bottom: var(--ref-space-8);
  flex-shrink: 0;
  background: transparent;
  border-bottom: 1px solid var(--ref-color-white-6);
}

.query-toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.query-toolbar-divider {
  width: 1px;
  height: 20px;
  background: var(--sys-color-divider-default);
}

.run-btn {
  border-radius: var(--ref-radius-md) !important;
  font-weight: var(--ref-font-weight-medium);
}

.query-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.query-hint {
  color: var(--ref-color-orange-6);
  font-size: var(--ref-font-size-xs);
}

.query-status-ok {
  color: var(--ref-color-green-5);
  font-size: var(--ref-font-size-xs);
  font-variant-numeric: tabular-nums;
}

.query-editor-area {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.query-splitter {
  height: var(--ref-space-5);
  cursor: row-resize;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.query-splitter::after {
  content: '';
  position: absolute;
  left: 30%;
  right: 30%;
  top: 2px;
  height: 1px;
  background: var(--ref-color-white-12);
  border-radius: var(--ref-radius-xs);
}

.query-splitter:hover::after {
  background: rgba(87, 114, 255, 0.4);
}

.query-results {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.query-no-rows {
  padding: var(--ref-space-24);
  text-align: center;
  color: var(--sys-color-text-tertiary);
  font-size: var(--ref-font-size-sm);
}

/* Welcome state */
.welcome-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: var(--ref-space-16);
  background: transparent;
}

.welcome-icon {
  opacity: 0.6;
}

.welcome-title {
  font-size: var(--ref-font-size-xl);
  font-weight: var(--ref-font-weight-semibold);
  color: var(--sys-color-text-disabled);
  letter-spacing: 1px;
}

.welcome-hint {
  font-size: var(--ref-font-size-sm);
  color: var(--ref-color-white-20);
}

/* =============================================
   Table Data Tab Layout
   ============================================= */
.data-tab-layout {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  overflow: hidden;
}

.data-spin-wrap {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ── Bottom Action Bar ── */
.data-bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  flex-shrink: 0;
  border-top: 1px solid #303640;
  background: rgba(255, 255, 255, 0.1);
}

.bottom-bar-left,
.bottom-bar-right {
  display: flex;
  align-items: center;
  gap: 4px;
}


.bar-divider {
  width: 1px;
  height: 18px;
  background: #303640;
  margin: 0 6px;
}

.bar-page-info {
  font-size: 12px;
  color: #A8ADB7;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.page-jump-input {
  width: 36px;
  height: 22px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  background: transparent;
  color: #FFFFFF;
  font-size: 12px;
  text-align: center;
  font-variant-numeric: tabular-nums;
  outline: none;
  padding: 0 2px;
  transition: border-color 0.15s ease;
}
.page-jump-input:focus {
  border-color: var(--sys-color-border-focus);
}

.bar-stats {
  font-size: 11px;
  color: #6B7280;
  white-space: nowrap;
}

/* ── Cell Editing ── */
.cell-edit-input {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  background: transparent;
  color: #FFFFFF;
  font-family: var(--ref-font-family-mono);
  font-size: var(--ref-font-size-xs);
  padding: 2px 4px;
  box-sizing: border-box;
}

.ptd-editing {
  padding: 0 !important;
  overflow: visible !important;
  outline: 2px solid var(--sys-color-border-focus);
  outline-offset: -2px;
  background: transparent !important;
}

.ptd-modified {
  box-shadow: inset 3px 0 0 rgba(255, 200, 0, 0.7);
}

.ptd-checkbox {
  width: 36px;
  text-align: center;
  padding: 0 !important;
}
.ptd-checkbox input[type="checkbox"] {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  width: 14px;
  height: 14px;
  border: 1.5px solid rgba(255, 255, 255, 0.25);
  border-radius: 3px;
  background: transparent;
  cursor: pointer;
  position: relative;
  vertical-align: middle;
  transition: background 0.15s ease, border-color 0.15s ease;
}
.ptd-checkbox input[type="checkbox"]:checked {
  background: var(--sys-color-border-focus);
  border-color: var(--sys-color-border-focus);
}
.ptd-checkbox input[type="checkbox"]:checked::after {
  content: '';
  position: absolute;
  left: 3.5px;
  top: 1px;
  width: 4px;
  height: 8px;
  border: solid #fff;
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
}

/* ── Row States ── */
.ptr-edited {
  box-shadow: inset 3px 0 0 rgba(255, 200, 0, 0.6);
}
/* Active row (clicked, non-edit mode visual focus) */
.ptr-active {
  background: rgba(91, 99, 246, 0.08) !important;
  box-shadow: inset 3px 0 0 var(--ref-color-brand-5, #5B63F6);
}
.ptr-active:hover {
  background: rgba(91, 99, 246, 0.12) !important;
}
.ptr-selected {
  background: rgba(91, 99, 246, 0.06) !important;
}
.ptr-deleted {
  text-decoration: line-through;
  opacity: 0.4;
  background: rgba(255, 69, 58, 0.05) !important;
}
.ptr-new {
  box-shadow: inset 3px 0 0 rgba(52, 199, 89, 0.7);
  background: rgba(52, 199, 89, 0.04) !important;
}

/* ── Structure Tab Layout ── */
.structure-tab-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 0;
}

.structure-content-shell {
  display: flex;
  flex: 1;
  min-height: 0;
  gap: var(--ref-space-12);
  align-items: stretch;
}

.table-inspector {
  width: 300px;
  min-width: 300px;
  flex: 0 0 300px;
  display: flex;
  flex-direction: column;
  gap: var(--ref-space-12);
  overflow: auto;
  padding-right: 2px;
}

.inspector-card {
  border: 1px solid var(--depth-1-border);
  border-radius: var(--ref-radius-xl);
  background: var(--depth-1-bg);
  padding: var(--ref-space-14);
}

.inspector-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ref-space-8);
  margin-bottom: var(--ref-space-10);
}

.inspector-section-title {
  color: var(--sys-color-text-title);
  font-size: var(--ref-font-size-sm);
  font-weight: var(--ref-font-weight-medium);
}

.inspector-kv-list {
  display: flex;
  flex-direction: column;
  gap: var(--ref-space-8);
}

.inspector-kv {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--ref-space-10);
  font-size: var(--ref-font-size-sm);
}

.inspector-kv span {
  color: var(--sys-color-text-secondary);
}

.inspector-kv strong {
  color: var(--sys-color-text-primary);
  font-weight: var(--ref-font-weight-medium);
  text-align: right;
  word-break: break-word;
}

.inspector-ddl {
  margin: 0;
  min-height: 220px;
  max-height: 360px;
  overflow: auto;
  padding: var(--ref-space-12);
  border: 1px solid var(--ref-color-white-10);
  border-radius: var(--ref-radius-lg);
  background: rgba(9, 11, 18, 0.55);
  color: var(--sys-color-text-primary);
  font-family: var(--ref-font-family-mono);
  font-size: var(--ref-font-size-xs);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.inspector-actions {
  margin-top: var(--ref-space-12);
  display: flex;
  gap: var(--ref-space-8);
  justify-content: flex-end;
}

.data-spin-wrap :deep(.n-spin-container) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.data-spin-wrap :deep(.n-spin-content) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ── Form View (Single Record Mode) ── */
.form-view-wrap {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.form-view-nav {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

.form-view-counter {
  /* Text/Body/SM */
  font-family: var(--ref-font-family-base);
  font-weight: var(--ref-font-weight-regular);
  font-size: var(--ref-font-size-sm);
  line-height: var(--ref-font-line-height-sm);
  color: var(--sys-color-text-secondary);
  font-variant-numeric: tabular-nums;
  min-width: 60px;
  text-align: center;
}

.form-view-fields {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 4px 0;
  scrollbar-width: thin;
  scrollbar-color: var(--ref-color-white-12) transparent;
}
.form-view-fields::-webkit-scrollbar { width: 6px; }
.form-view-fields::-webkit-scrollbar-track { background: transparent; }
.form-view-fields::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.12); border-radius: 3px; }

.form-view-field {
  display: flex;
  min-height: 36px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  transition: background 0.12s ease;
}
.form-view-field:hover {
  background: rgba(255, 255, 255, 0.03);
}

.form-view-label {
  /* Text/Label/MD */
  flex-shrink: 0;
  width: 180px;
  padding: var(--ref-space-8) var(--ref-space-14);
  font-family: var(--ref-font-family-base);
  font-weight: var(--ref-font-weight-medium);
  font-size: var(--ref-font-size-md);
  line-height: var(--ref-font-line-height-md);
  color: var(--sys-color-text-secondary);
  word-break: break-all;
  border-right: 1px solid var(--ref-color-white-6);
  display: flex;
  align-items: flex-start;
}

.form-view-value {
  /* Text/Body/SM */
  flex: 1;
  min-width: 0;
  padding: var(--ref-space-8) var(--ref-space-14);
  font-family: var(--ref-font-family-base);
  font-weight: var(--ref-font-weight-regular);
  font-size: var(--ref-font-size-sm);
  line-height: var(--ref-font-line-height-sm);
  color: var(--sys-color-text-primary);
}

.form-view-pre {
  /* Text/Body/SM — inherits, only resets margin & whitespace */
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  font-weight: inherit;
  line-height: inherit;
  white-space: pre-wrap;
  word-break: break-all;
  color: inherit;
}

.form-view-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--sys-color-text-tertiary);
  font-family: var(--ref-font-family-base);
  font-size: var(--ref-font-size-sm);
}

/* Form view: editable textarea */
.form-view-input {
  display: block;
  width: 100%;
  min-height: 28px;
  padding: var(--ref-space-4) var(--ref-space-8);
  border: 1px solid var(--ref-color-white-12);
  border-radius: var(--ref-radius-sm);
  background: var(--sys-color-bg-form-control);
  color: var(--sys-color-text-primary);
  font-family: var(--ref-font-family-base);
  font-weight: var(--ref-font-weight-regular);
  font-size: var(--ref-font-size-sm);
  line-height: var(--ref-font-line-height-sm);
  resize: vertical;
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.form-view-input:focus {
  border-color: var(--sys-color-border-focus);
  box-shadow: var(--ref-shadow-focus-brand);
}

/* Modified field indicator */
.form-view-field--modified {
  box-shadow: inset 3px 0 0 rgba(255, 200, 0, 0.6);
  background: rgba(255, 214, 0, 0.03);
}

/* Filter bar */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 0 7px;
  flex-shrink: 0;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  margin-bottom: 4px;
}
.filter-label {
  font-size: 11px;
  font-weight: 700;
  color: rgba(87,114,255,0.8);
  font-family: monospace;
  flex-shrink: 0;
}
.filter-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(87,114,255,0.15);
  color: rgba(87,114,255,0.9);
  border-radius: 4px;
  padding: 1px 6px;
  font-size: 11px;
  cursor: pointer;
  margin-left: 4px;
}
.filter-badge:hover { background: rgba(87,114,255,0.25); }

/* Edit action bar */
.edit-action-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  background: rgba(255,214,0,0.06);
  border: 1px solid rgba(255,214,0,0.2);
  border-radius: 6px;
  font-size: 12px;
  color: rgba(255,214,0,0.8);
  flex-shrink: 0;
  margin-bottom: 6px;
}
.edit-action-bar span { flex: 1; }

@media (max-width: 1280px) {
  .table-inspector {
    width: 280px;
    min-width: 280px;
    flex-basis: 280px;
  }
}


/* =============================================
   Premium Query Results Table
   ============================================= */

/* Scrollable wrapper */
.premium-table-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  background: transparent;
  scrollbar-width: thin;
  scrollbar-color: var(--ref-color-white-12) transparent;
}
.premium-table-wrap::-webkit-scrollbar { width: var(--scroll-size-default); height: var(--scroll-size-default); }
.premium-table-wrap::-webkit-scrollbar-track { background: var(--scroll-track-bg); }
.premium-table-wrap::-webkit-scrollbar-thumb { background: var(--scroll-thumb-bg); border-radius: var(--scroll-thumb-radius); }
.premium-table-wrap::-webkit-scrollbar-thumb:hover { background: var(--scroll-thumb-bg-hover); }

/* Table itself */
.premium-table {
  width: max-content;
  min-width: 100%;
  border-collapse: collapse;
  font-size: var(--ref-font-size-sm);
  line-height: 1.5;
  background: transparent;
}

/* ---- Sticky Header: uses global .gm-pth ---- */
/* Component-specific header add-ons only */

/* Sortable column header */
.pth-sortable {
  cursor: pointer;
  user-select: none;
}
.pth-sortable:hover { background: rgba(87,114,255,0.08) !important; }

.th-content {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: space-between;
}
.sort-indicator {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

/* ---- Data Rows: uses global .gm-ptr ---- */
/* Component-specific row add-ons only */

/* Zebra stripe */
.ptr-even { background: rgba(255, 255, 255, 0.012); }

/* Edited row (has pending changes) */
.ptr-edited {
  box-shadow: inset 3px 0 0 rgba(255, 214, 0, 0.7) !important;
  background: rgba(255, 214, 0, 0.04) !important;
}

/* ---- Cells: uses global .gm-ptd ---- */
/* Component-specific cell add-ons only */

/* ── Schema-specific cell helpers ── */
.ptd-mono {
  font-family: var(--ref-font-family-mono);
  font-size: var(--ref-font-size-xs);
  color: rgba(165, 214, 255, 0.85);
}
.ptd-name {
  font-weight: var(--ref-font-weight-semibold);
  color: var(--sys-color-text-title);
}
.ptd-center { text-align: center; }
.ptd-comment {
  color: var(--sys-color-text-secondary);
  font-style: italic;
}

.schema-badge {
  display: inline-block;
  padding: 1px 7px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: var(--ref-font-weight-medium);
  letter-spacing: 0.2px;
}
.schema-badge--yes {
  background: rgba(82, 196, 100, 0.12);
  color: var(--ref-color-green-5);
  border: 1px solid rgba(82, 196, 100, 0.2);
}
.schema-badge--no {
  background: rgba(255, 255, 255, 0.05);
  color: var(--sys-color-text-tertiary);
  border: 1px solid var(--ref-color-white-8);
}

.cell-content {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* NULL badge */
.null-badge {
  display: inline-block;
  color: var(--sys-color-text-tertiary);
  font-size: var(--ref-font-size-sm);
  font-family: inherit;
  font-style: normal;
  font-weight: var(--ref-font-weight-regular);
  letter-spacing: normal;
}

.hint-badge {
  display: inline-block;
  color: rgba(255, 255, 255, 0.3);
  font-size: var(--ref-font-size-xs);
  font-style: italic;
  line-height: 1.4;
  white-space: nowrap;
  overflow: visible;
}

/* ---- Breathing LED in status bar ---- */
.status-led {
  display: inline-block;
  width: var(--ref-space-6);
  height: var(--ref-space-6);
  border-radius: 50%;
  background: var(--ref-color-green-5);
  box-shadow: 0 0 6px var(--ref-color-green-5);
  margin-right: var(--ref-space-6);
  vertical-align: middle;
  animation: led-pulse 2.4s ease-in-out infinite;
}
@keyframes led-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 6px var(--ref-color-green-5); }
  50%       { opacity: 0.5; box-shadow: 0 0 2px var(--ref-color-green-5); }
}

/* ---- Premium Empty State ---- */
.premium-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: var(--ref-space-48) var(--ref-space-24);
}
.empty-svg {
  width: 100px;
  height: 84px;
  opacity: 0.9;
  animation: float 4s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50%       { transform: translateY(-6px); }
}
.empty-title {
  font-size: var(--ref-font-size-md);
  font-weight: var(--ref-font-weight-semibold);
  color: var(--ref-color-white-25);
  letter-spacing: 0.5px;
}
.empty-sub {
  font-size: var(--ref-font-size-xs);
  color: var(--ref-color-white-15);
}
</style>
