import { useEffect, useRef, useState } from 'react'

import {
  DEFAULT_PLAYER_HOTKEYS,
  PLAYER_HOTKEY_ACTIONS,
  formatPlayerHotkeyKey,
  isReservedPlayerHotkeyKey,
  isSupportedPlayerHotkeyKey,
  normalizePlayerHotkey,
  normalizePlayerHotkeyKey,
} from '@/utils/playerHotkeys'
import { zh } from '@/utils/i18n'

const createRows = (items) =>
  (Array.isArray(items) && items.length ? items : DEFAULT_PLAYER_HOTKEYS).map((item, index) => ({
    id: `${Date.now()}-${index}-${item.key}`,
    key: item.key,
    action: item.action,
    amount: String(item.amount),
  }))

export default function PlayerSettingsModal({ open, onClose, hotkeys, onSave }) {
  const seedRef = useRef(0)
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    seedRef.current += 1
    const source = Array.isArray(hotkeys) ? hotkeys : DEFAULT_PLAYER_HOTKEYS
    setRows(createRows(source))
    setError('')
    setSaving(false)
  }, [open, hotkeys])

  if (!open) return null

  const setRowValue = (id, patch) => {
    setRows((current) => current.map((row) => (row.id === id ? { ...row, ...patch } : row)))
  }

  const handleAdd = () => {
    seedRef.current += 1
    setRows((current) => [
      ...current,
      {
        id: `${seedRef.current}-new`,
        key: '',
        action: PLAYER_HOTKEY_ACTIONS.SEEK,
        amount: '5',
      },
    ])
    setError('')
  }

  const handleReset = () => {
    setRows(createRows(DEFAULT_PLAYER_HOTKEYS))
    setError('')
  }

  const handleSave = async () => {
    setError('')
    const seen = new Set()
    const normalized = []
    for (const row of rows) {
      const key = normalizePlayerHotkeyKey(row.key)
      if (!key) {
        setError(zh('请为每个快捷键设置按键', 'Set a key for every shortcut'))
        return
      }
      if (!isSupportedPlayerHotkeyKey(key)) {
        setError(
          zh(
            `按键 ${formatPlayerHotkeyKey(key)} 不支持`,
            `Key ${formatPlayerHotkeyKey(key)} is not supported`
          )
        )
        return
      }
      if (isReservedPlayerHotkeyKey(key)) {
        setError(
          zh(
            `按键 ${formatPlayerHotkeyKey(key)} 已保留给播放/关闭`,
            `Key ${formatPlayerHotkeyKey(key)} is reserved for play/close`
          )
        )
        return
      }
      if (seen.has(key)) {
        setError(
          zh(
            `按键 ${formatPlayerHotkeyKey(key)} 重复了`,
            `Key ${formatPlayerHotkeyKey(key)} is duplicated`
          )
        )
        return
      }
      seen.add(key)
      const item = normalizePlayerHotkey({
        key,
        action: row.action,
        amount: row.amount,
      })
      if (!item) {
        setError(
          zh(
            `请检查按键 ${formatPlayerHotkeyKey(key)} 的动作和数值`,
            `Check the action and value for key ${formatPlayerHotkeyKey(key)}`
          )
        )
        return
      }
      normalized.push(item)
    }

    setSaving(true)
    try {
      await onSave?.(normalized)
      onClose?.()
    } catch (err) {
      setError(err.message || zh('保存失败', 'Save failed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 px-4">
      <div className="w-full max-w-4xl rounded-lg bg-white p-5 shadow-xl">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold">{zh('播放器设置', 'Player Settings')}</h2>
            <p className="mt-1 text-sm text-gray-500">
              {zh('仅支持视频进度和音量快捷键', 'Only seek and volume shortcuts are supported')}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100"
          >
            {zh('关闭', 'Close')}
          </button>
        </div>

        <div className="mt-4 space-y-3">
          <div className="grid grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] gap-3 px-2 text-xs font-medium uppercase tracking-wide text-gray-500">
            <div>{zh('按键', 'Key')}</div>
            <div>{zh('动作', 'Action')}</div>
            <div>{zh('变化值', 'Amount')}</div>
            <div>{zh('操作', 'Operation')}</div>
          </div>

          {rows.length === 0 ? (
            <div className="rounded border border-dashed px-4 py-6 text-center text-sm text-gray-500">
              {zh('当前没有自定义快捷键', 'No custom shortcuts')}
            </div>
          ) : (
            rows.map((row) => (
              <div
                key={row.id}
                className="grid grid-cols-1 gap-3 rounded border p-3 md:grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]"
              >
                <input
                  value={formatPlayerHotkeyKey(row.key)}
                  readOnly
                  onKeyDown={(event) => {
                    event.preventDefault()
                    event.stopPropagation()
                    const nextKey = normalizePlayerHotkeyKey(event.key)
                    if (!nextKey) return
                    setRowValue(row.id, { key: nextKey })
                    setError('')
                  }}
                  onFocus={() => setError('')}
                  placeholder={zh('聚焦后按下按键', 'Focus then press a key')}
                  className="rounded border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <select
                  value={row.action}
                  onChange={(event) => {
                    const nextAction = event.target.value
                    setRowValue(row.id, {
                      action: nextAction,
                      amount: nextAction === PLAYER_HOTKEY_ACTIONS.VOLUME ? '10' : '5',
                    })
                    setError('')
                  }}
                  className="rounded border px-3 py-2 text-sm"
                >
                  <option value={PLAYER_HOTKEY_ACTIONS.SEEK}>{zh('调节进度', 'Seek')}</option>
                  <option value={PLAYER_HOTKEY_ACTIONS.VOLUME}>{zh('调节音量', 'Volume')}</option>
                </select>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    step="1"
                    value={row.amount}
                    onChange={(event) => {
                      setRowValue(row.id, { amount: event.target.value })
                      setError('')
                    }}
                    className="w-full rounded border px-3 py-2 text-sm"
                  />
                  <span className="shrink-0 text-sm text-gray-500">
                    {row.action === PLAYER_HOTKEY_ACTIONS.VOLUME ? '%' : zh('秒', 'sec')}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    setRows((current) => current.filter((item) => item.id !== row.id))
                    setError('')
                  }}
                  className="rounded border px-3 py-2 text-sm text-red-600 hover:bg-red-50"
                >
                  {zh('删除', 'Delete')}
                </button>
              </div>
            ))
          )}
        </div>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={handleAdd}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              {zh('新增快捷键', 'Add shortcut')}
            </button>
            <button
              type="button"
              onClick={handleReset}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              {zh('恢复默认', 'Restore defaults')}
            </button>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              {zh('取消', 'Cancel')}
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {saving ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
            </button>
          </div>
        </div>

        <div className="mt-2 text-xs text-gray-500">
          {zh(
            '正数表示增加，负数表示减少。`Space` 和 `Escape` 仍固定用于播放/暂停和关闭播放器。',
            'Positive numbers increase, negative numbers decrease. `Space` and `Escape` remain reserved for play/pause and close.'
          )}
        </div>
        {error && <div className="mt-2 text-sm text-red-600">{error}</div>}
      </div>
    </div>
  )
}
