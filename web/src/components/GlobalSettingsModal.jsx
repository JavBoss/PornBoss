import { useEffect, useMemo, useState } from 'react'

import DirectoryManager from '@/components/DirectoryManager'
import PlayerSettingsModal from '@/components/PlayerSettingsModal'
import { formatPlayerHotkeyKey, parsePlayerHotkeys } from '@/utils/playerHotkeys'
import { zh } from '@/utils/i18n'

const SETTINGS_SECTIONS = [
  {
    id: 'proxy',
    title: { zh: '网络与代理', en: 'Network & Proxy' },
    summary: { zh: '代理端口与连接行为', en: 'Proxy port and connection behavior' },
  },
  {
    id: 'player',
    title: { zh: '播放器设置', en: 'Player Settings' },
    summary: { zh: 'mpv 快捷键与播放控制', en: 'mpv shortcuts and playback controls' },
  },
  {
    id: 'directories',
    title: { zh: '目录管理', en: 'Directory Management' },
    summary: { zh: '管理扫描目录与路径', en: 'Manage watched folders and paths' },
  },
]

export default function GlobalSettingsModal({
  open,
  onClose,
  directories,
  onCreateDirectory,
  onUpdateDirectory,
  onDeleteDirectory,
  proxyPort,
  onSaveProxyPort,
  playerHotkeys,
  onSavePlayerHotkeys,
}) {
  const [proxyInput, setProxyInput] = useState('')
  const [proxyError, setProxyError] = useState('')
  const [savingProxy, setSavingProxy] = useState(false)
  const [proxyEditing, setProxyEditing] = useState(false)
  const [proxyEnabledInput, setProxyEnabledInput] = useState(false)
  const [activeSection, setActiveSection] = useState('proxy')

  const normalizedPlayerHotkeys = useMemo(() => parsePlayerHotkeys(playerHotkeys), [playerHotkeys])
  const shortcutSummary = normalizedPlayerHotkeys
    .map((item) =>
      item.action === 'volume'
        ? zh(
            `${formatPlayerHotkeyKey(item.key)}: 音量 ${item.amount > 0 ? '+' : ''}${item.amount}%`,
            `${formatPlayerHotkeyKey(item.key)}: Volume ${item.amount > 0 ? '+' : ''}${item.amount}%`
          )
        : zh(
            `${formatPlayerHotkeyKey(item.key)}: 进度 ${item.amount > 0 ? '+' : ''}${item.amount} 秒`,
            `${formatPlayerHotkeyKey(item.key)}: Seek ${item.amount > 0 ? '+' : ''}${item.amount}s`
          )
    )
    .join(' / ')

  useEffect(() => {
    if (open) {
      setProxyInput(proxyPort ? String(proxyPort) : '')
      setProxyEnabledInput(Boolean(proxyPort))
      setProxyEditing(false)
      setProxyError('')
    }
  }, [open, proxyPort])

  if (!open) return null

  const handleSaveProxy = async () => {
    setProxyError('')
    const raw = proxyInput.trim()
    let port = 0
    if (proxyEnabledInput) {
      if (raw === '') {
        setProxyError(zh('请输入 1-65535 的端口号', 'Enter a port between 1 and 65535'))
        return
      }
      const parsed = parseInt(raw, 10)
      if (!Number.isFinite(parsed) || parsed <= 0 || parsed > 65535) {
        setProxyError(zh('请输入 1-65535 的端口号', 'Enter a port between 1 and 65535'))
        return
      }
      port = parsed
    }
    setSavingProxy(true)
    try {
      await onSaveProxyPort?.(port)
      setProxyEditing(false)
    } catch (err) {
      setProxyError(err.message || zh('保存失败', 'Save failed'))
    } finally {
      setSavingProxy(false)
    }
  }

  const proxyInputTrimmed = proxyInput.trim()
  const desiredPortText = proxyEnabledInput ? proxyInputTrimmed : ''
  const currentPortText = proxyPort ? String(proxyPort) : ''
  const proxyUnchanged = desiredPortText === currentPortText
  const proxyInputMissing = proxyEnabledInput && proxyInputTrimmed === ''
  const activeTitle = SETTINGS_SECTIONS.find((item) => item.id === activeSection)?.title || {
    zh: '全局设置',
    en: 'Global Settings',
  }

  const renderProxyPanel = () => (
    <div className="space-y-5">
      <section className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h4 className="text-sm font-semibold text-zinc-800">
                {zh('代理端口', 'Proxy Port')}
              </h4>
              <p className="mt-1 text-sm text-zinc-500">
                {proxyPort
                  ? zh(`当前使用端口 ${proxyPort}`, `Currently using port ${proxyPort}`)
                  : zh('当前使用自动检测', 'Currently using auto-detection')}
              </p>
            </div>
            {!proxyEditing && (
              <button
                type="button"
                onClick={() => {
                  setProxyEditing(true)
                  setProxyError('')
                }}
                className="rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-sm text-zinc-700 hover:bg-zinc-50"
              >
                {zh('编辑', 'Edit')}
              </button>
            )}
          </div>

          {proxyEditing ? (
            <div className="space-y-4 rounded-2xl bg-zinc-50 p-4">
              <label className="flex items-center gap-2 text-sm text-zinc-700">
                <input
                  type="checkbox"
                  checked={proxyEnabledInput}
                  onChange={(e) => {
                    setProxyEnabledInput(e.target.checked)
                    setProxyError('')
                  }}
                  className="h-4 w-4 rounded"
                />
                <span>{zh('手动设置端口', 'Set port manually')}</span>
              </label>

              {proxyEnabledInput && (
                <div className="max-w-sm">
                  <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-zinc-500">
                    {zh('端口号', 'Port')}
                  </label>
                  <input
                    value={proxyInput}
                    onChange={(e) => setProxyInput(e.target.value)}
                    placeholder={zh('输入 1-65535', 'Enter 1-65535')}
                    inputMode="numeric"
                    className="w-full rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm"
                  />
                </div>
              )}

              {proxyError && <div className="text-sm text-red-600">{proxyError}</div>}

              <div className="flex flex-wrap justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setProxyInput(proxyPort ? String(proxyPort) : '')
                    setProxyEnabledInput(Boolean(proxyPort))
                    setProxyError('')
                    setProxyEditing(false)
                  }}
                  className="rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-sm text-zinc-700 hover:bg-zinc-50"
                >
                  {zh('取消', 'Cancel')}
                </button>
                <button
                  type="button"
                  onClick={handleSaveProxy}
                  disabled={savingProxy || proxyUnchanged || proxyInputMissing}
                  className="rounded-xl bg-blue-600 px-3 py-1.5 text-sm text-white disabled:opacity-60"
                >
                  {savingProxy ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
                </button>
              </div>
            </div>
          ) : null}
        </div>
      </section>
    </div>
  )

  const renderPlayerPanel = () => (
    <div className="space-y-5">
      <section className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
        <div className="mb-4 text-sm text-zinc-500">
          {shortcutSummary || zh('当前未配置自定义快捷键', 'No shortcuts configured')}
        </div>
        <PlayerSettingsModal
          embedded
          hotkeys={normalizedPlayerHotkeys}
          onSave={onSavePlayerHotkeys}
        />
      </section>
    </div>
  )

  const renderDirectoriesPanel = () => (
    <div className="space-y-5">
      <section className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
        <DirectoryManager
          open={open}
          directories={directories}
          onCreate={onCreateDirectory}
          onUpdate={onUpdateDirectory}
          onDelete={onDeleteDirectory}
        />
      </section>
    </div>
  )

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <div className="flex h-[min(86vh,820px)] w-full max-w-6xl flex-col overflow-hidden rounded-[28px] border border-zinc-200 bg-[#f5f5f7] shadow-2xl">
        <div className="flex items-center justify-between border-b border-zinc-200 bg-white/70 px-6 py-4 backdrop-blur">
          <div>
            <h2 className="text-lg font-semibold text-zinc-900">
              {zh('全局设置', 'Global Settings')}
            </h2>
            <p className="mt-1 text-sm text-zinc-500">{zh(activeTitle.zh, activeTitle.en)}</p>
          </div>
          <button
            onClick={onClose}
            className="rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-50"
          >
            {zh('关闭', 'Close')}
          </button>
        </div>

        <div className="flex min-h-0 flex-1 flex-col md:flex-row">
          <aside className="border-b border-zinc-200 bg-white/60 p-3 backdrop-blur md:w-[280px] md:border-b-0 md:border-r">
            <div className="flex gap-2 overflow-x-auto md:flex-col">
              {SETTINGS_SECTIONS.map((section) => {
                const selected = activeSection === section.id
                const badgeText =
                  section.id === 'proxy'
                    ? proxyPort
                      ? String(proxyPort)
                      : zh('自动', 'Auto')
                    : section.id === 'player'
                      ? String(normalizedPlayerHotkeys.length)
                      : String(directories.length)

                return (
                  <button
                    key={section.id}
                    type="button"
                    onClick={() => setActiveSection(section.id)}
                    className={`min-w-[220px] rounded-2xl border px-4 py-3 text-left transition md:min-w-0 ${
                      selected
                        ? 'border-zinc-200 bg-white shadow-sm'
                        : 'border-transparent bg-transparent hover:border-zinc-200 hover:bg-white/80'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="text-sm font-semibold text-zinc-900">
                          {zh(section.title.zh, section.title.en)}
                        </div>
                        <div className="mt-1 text-xs text-zinc-500">
                          {zh(section.summary.zh, section.summary.en)}
                        </div>
                      </div>
                      <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-600">
                        {badgeText}
                      </span>
                    </div>
                  </button>
                )
              })}
            </div>
          </aside>

          <section className="min-h-0 flex-1 overflow-y-auto px-4 py-4 md:px-6 md:py-6">
            {activeSection === 'proxy' && renderProxyPanel()}
            {activeSection === 'player' && renderPlayerPanel()}
            {activeSection === 'directories' && renderDirectoriesPanel()}
          </section>
        </div>
      </div>
    </div>
  )
}
