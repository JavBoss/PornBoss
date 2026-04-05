import { useEffect, useMemo, useState } from 'react'
import CloseOutlinedIcon from '@mui/icons-material/CloseOutlined'

import TagBar from '@/components/TagBar'
import { zh } from '@/utils/i18n'

export default function VideoTagModal({
  open,
  onClose,
  tags,
  onToggleFilter,
  onCreateTag,
  onDeleteTag,
  onRenameTag,
  onApplyTagFilter,
}) {
  const [createOpen, setCreateOpen] = useState(false)
  const [newTagName, setNewTagName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')
  const [renameOpen, setRenameOpen] = useState(false)
  const [renameTagId, setRenameTagId] = useState(null)
  const [renameOriginalName, setRenameOriginalName] = useState('')
  const [renameTagName, setRenameTagName] = useState('')
  const [renaming, setRenaming] = useState(false)
  const [renameError, setRenameError] = useState('')
  const [editMode, setEditMode] = useState(false)
  const [hoverTagId, setHoverTagId] = useState(null)
  const [multiSelect, setMultiSelect] = useState(false)
  const [selectedTagIds, setSelectedTagIds] = useState([])
  const [batchError, setBatchError] = useState('')
  const [deletingId, setDeletingId] = useState(null)

  const handleTagClick = (name) => {
    if (typeof onApplyTagFilter === 'function') {
      onApplyTagFilter([name])
      onClose()
      return
    }
    onToggleFilter?.(name)
    onClose()
  }

  useEffect(() => {
    if (!open) {
      setCreateOpen(false)
      setNewTagName('')
      setCreateError('')
      setCreating(false)
      setRenameOpen(false)
      setRenameTagId(null)
      setRenameOriginalName('')
      setRenameTagName('')
      setRenaming(false)
      setRenameError('')
      setEditMode(false)
      setHoverTagId(null)
      setMultiSelect(false)
      setSelectedTagIds([])
      setBatchError('')
      setDeletingId(null)
    }
  }, [open])

  const selectedTags = useMemo(() => {
    if (selectedTagIds.length === 0) return []
    const set = new Set(selectedTagIds)
    return tags.filter((t) => set.has(t.id))
  }, [selectedTagIds, tags])

  const selectedNames = useMemo(() => selectedTags.map((t) => t.name), [selectedTags])
  const handleStartRename = (tag) => {
    setRenameOpen(true)
    setRenameTagId(tag.id)
    setRenameOriginalName(tag.name || '')
    setRenameTagName(tag.name || '')
    setRenameError('')
    setRenaming(false)
  }
  const handleCloseRename = () => {
    setRenameOpen(false)
    setRenameTagId(null)
    setRenameOriginalName('')
    setRenameTagName('')
    setRenameError('')
    setRenaming(false)
  }
  const handleToggleEditMode = () => {
    setEditMode((prev) => !prev)
    setHoverTagId(null)
  }

  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 backdrop-blur-sm">
      <div className="mx-4 w-full max-w-4xl overflow-hidden rounded-3xl bg-white shadow-2xl ring-1 ring-slate-200/70">
        <div className="flex items-center justify-between border-b border-slate-200/70 bg-slate-50/80 px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">
            {zh('标签管理', 'Tag Management')}
          </h2>
          <button
            onClick={onClose}
            className="rounded-full bg-slate-100 px-3 py-1.5 text-sm font-medium text-slate-600 transition hover:bg-slate-200"
            aria-label={zh('关闭', 'Close')}
          >
            {zh('关闭', 'Close')}
          </button>
        </div>
        <div className="space-y-6 p-6">
          <section className="space-y-4">
            <div className="max-h-[65vh] overflow-y-auto pr-1">
              {multiSelect ? (
                <TagBar
                  tags={tags}
                  onToggle={handleTagClick}
                  multiSelect={multiSelect}
                  selectedIds={selectedTagIds}
                  onSelect={(id) => {
                    setSelectedTagIds((prev) => {
                      const next = new Set(prev)
                      if (next.has(id)) {
                        next.delete(id)
                      } else {
                        next.add(id)
                      }
                      return Array.from(next)
                    })
                  }}
                />
              ) : (
                <div className="flex flex-wrap gap-2">
                  {tags.map((t) => {
                    const count = Number.isFinite(t.count) ? t.count : null
                    const showRenameHint = editMode && hoverTagId === t.id
                    const showDelete = editMode && hoverTagId === t.id
                    return (
                      <div
                        key={t.id}
                        className="inline-flex items-center gap-2 rounded px-2 py-1 text-sm text-slate-700 transition hover:bg-slate-100"
                        onMouseEnter={() => {
                          if (editMode) setHoverTagId(t.id)
                        }}
                        onMouseLeave={() => {
                          if (editMode) setHoverTagId((prev) => (prev === t.id ? null : prev))
                        }}
                      >
                        <button
                          type="button"
                          className="flex items-center gap-2 text-left"
                          onClick={() => {
                            if (editMode) {
                              handleStartRename(t)
                              return
                            }
                            handleTagClick(t.name)
                          }}
                          title={t.name}
                        >
                          <span>{t.name}</span>
                          {!editMode && count !== null && (
                            <span className="rounded-full bg-slate-100 px-1.5 text-[10px] text-slate-500">
                              {count}
                            </span>
                          )}
                          {showRenameHint && (
                            <span className="text-[10px] text-slate-400">
                              {zh('单击重命名', 'Click to rename')}
                            </span>
                          )}
                        </button>
                        {showDelete && (
                          <button
                            type="button"
                            aria-label={zh('删除标签', 'Delete tag')}
                            disabled={deletingId === t.id}
                            onClick={async (event) => {
                              event.preventDefault()
                              event.stopPropagation()
                              if (deletingId === t.id) return
                              if (
                                !window.confirm(
                                  zh(`确定删除标签“${t.name}”吗？`, `Delete tag "${t.name}"?`)
                                )
                              )
                                return
                              setDeletingId(t.id)
                              setBatchError('')
                              try {
                                await onDeleteTag?.(t)
                              } catch (err) {
                                setBatchError(err.message || zh('删除失败', 'Delete failed'))
                              } finally {
                                setDeletingId(null)
                              }
                            }}
                          >
                            <CloseOutlinedIcon
                              fontSize="inherit"
                              className="h-3.5 w-3.5 text-rose-500"
                            />
                          </button>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                {!editMode && (
                  <button
                    onClick={() => {
                      setBatchError('')
                      setMultiSelect((prev) => !prev)
                      setSelectedTagIds([])
                      setEditMode(false)
                      setHoverTagId(null)
                    }}
                    className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-100"
                  >
                    {multiSelect ? zh('退出多选', 'Exit multi-select') : zh('多选', 'Multi-select')}
                  </button>
                )}
                {!multiSelect && (
                  <>
                    <button
                      onClick={handleToggleEditMode}
                      className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-100"
                    >
                      {editMode ? zh('退出编辑', 'Exit edit') : zh('编辑', 'Edit')}
                    </button>
                    {!editMode && (
                      <button
                        onClick={() => {
                          setCreateError('')
                          setNewTagName('')
                          setCreateOpen(true)
                        }}
                        className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-100"
                      >
                        {zh('新增标签', 'New tag')}
                      </button>
                    )}
                  </>
                )}
                {multiSelect && (
                  <button
                    onClick={() => {
                      if (selectedNames.length === 0) return
                      onApplyTagFilter(selectedNames)
                      onClose()
                    }}
                    disabled={selectedNames.length === 0}
                    className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-100 disabled:opacity-60"
                  >
                    {zh('查找视频', 'Find videos')}
                  </button>
                )}
              </div>
            </div>
            {batchError && <div className="text-sm text-rose-600">{batchError}</div>}
          </section>
        </div>
      </div>
      {renameOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-base font-semibold text-slate-900">
                {zh('重命名标签', 'Rename tag')}
              </h3>
              <button
                onClick={handleCloseRename}
                className="rounded px-2 py-1 text-sm text-slate-500 hover:bg-slate-100"
                aria-label={zh('关闭重命名', 'Close rename')}
              >
                ✕
              </button>
            </div>
            <div className="space-y-3">
              <input
                value={renameTagName}
                onChange={(e) => setRenameTagName(e.target.value)}
                placeholder={zh('请输入新的标签名', 'Enter a new tag name')}
                className="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                onKeyDown={(event) => {
                  if (event.key === 'Escape') {
                    handleCloseRename()
                  }
                }}
              />
              {renameError && <div className="text-sm text-red-600">{renameError}</div>}
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button
                onClick={handleCloseRename}
                className="rounded border px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
              >
                {zh('取消', 'Cancel')}
              </button>
              <button
                onClick={async () => {
                  const trimmed = renameTagName.trim()
                  if (!trimmed) {
                    setRenameError(zh('标签名不能为空', 'Tag name cannot be empty'))
                    return
                  }
                  if (!renameTagId) {
                    setRenameError(zh('标签不存在', 'Tag not found'))
                    return
                  }
                  if (trimmed === renameOriginalName) {
                    handleCloseRename()
                    return
                  }
                  setRenaming(true)
                  setRenameError('')
                  try {
                    await onRenameTag?.(renameTagId, trimmed)
                    handleCloseRename()
                  } catch (err) {
                    setRenameError(err.message || zh('重命名失败', 'Rename failed'))
                  } finally {
                    setRenaming(false)
                  }
                }}
                disabled={renaming}
                className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-60"
              >
                {renaming ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
              </button>
            </div>
          </div>
        </div>
      )}
      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-base font-semibold text-slate-900">
                {zh('新增标签', 'New tag')}
              </h3>
              <button
                onClick={() => setCreateOpen(false)}
                className="rounded px-2 py-1 text-sm text-slate-500 hover:bg-slate-100"
                aria-label={zh('关闭新增标签', 'Close new tag')}
              >
                ✕
              </button>
            </div>
            <div className="space-y-3">
              <input
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder={zh('请输入标签名', 'Enter tag name')}
                className="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
              {createError && <div className="text-sm text-red-600">{createError}</div>}
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button
                onClick={() => setCreateOpen(false)}
                className="rounded border px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
              >
                {zh('取消', 'Cancel')}
              </button>
              <button
                onClick={async () => {
                  const trimmed = newTagName.trim()
                  if (!trimmed) {
                    setCreateError(zh('标签名不能为空', 'Tag name cannot be empty'))
                    return
                  }
                  setCreating(true)
                  setCreateError('')
                  try {
                    await onCreateTag(trimmed)
                    setCreateOpen(false)
                    setNewTagName('')
                  } catch (err) {
                    setCreateError(err.message || zh('创建失败', 'Create failed'))
                  } finally {
                    setCreating(false)
                  }
                }}
                disabled={creating}
                className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-60"
              >
                {creating ? zh('创建中…', 'Creating...') : zh('创建', 'Create')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
