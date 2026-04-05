import { useEffect, useState } from 'react'

import DirectoryManager from '@/components/DirectoryManager'
import PlayerSettingsModal from '@/components/PlayerSettingsModal'
import { zh } from '@/utils/i18n'

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
  const [playerSettingsOpen, setPlayerSettingsOpen] = useState(false)

  useEffect(() => {
    if (open) {
      setProxyInput(proxyPort ? String(proxyPort) : '')
      setProxyEnabledInput(Boolean(proxyPort))
      setProxyEditing(false)
      setProxyError('')
      setPlayerSettingsOpen(false)
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

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <div className="w-full max-w-4xl rounded-lg bg-white p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">{zh('全局设置', 'Global Settings')}</h2>
          <button onClick={onClose} className="rounded px-3 py-1 text-gray-600 hover:bg-gray-100">
            {zh('关闭', 'Close')}
          </button>
        </div>

        <div className="mt-4 grid gap-6">
          <section className="rounded-lg border p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h3 className="text-sm font-semibold text-gray-700">
                {zh('代理端口', 'Proxy Port')}
              </h3>
              {!proxyEditing ? (
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-500">
                    {proxyPort
                      ? zh(`端口：${proxyPort}`, `Port: ${proxyPort}`)
                      : zh('自动检测', 'Auto-detect')}
                  </span>
                  <button
                    type="button"
                    onClick={() => {
                      setProxyEditing(true)
                      setProxyError('')
                    }}
                    className="rounded border px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                  >
                    {zh('设置', 'Edit')}
                  </button>
                </div>
              ) : (
                <div className="flex flex-wrap items-center gap-2">
                  <label className="flex items-center gap-1 text-xs text-gray-600">
                    <span>{zh('手动设置端口', 'Set port manually')}</span>
                    <input
                      type="checkbox"
                      checked={proxyEnabledInput}
                      onChange={(e) => {
                        setProxyEnabledInput(e.target.checked)
                        setProxyError('')
                      }}
                      className="h-3.5 w-3.5"
                    />
                  </label>
                  {proxyEnabledInput && (
                    <input
                      value={proxyInput}
                      onChange={(e) => setProxyInput(e.target.value)}
                      placeholder={zh('端口号', 'Port')}
                      inputMode="numeric"
                      className="w-24 rounded border px-3 py-1.5 text-sm"
                    />
                  )}
                  <button
                    type="button"
                    onClick={handleSaveProxy}
                    disabled={savingProxy || proxyUnchanged || proxyInputMissing}
                    className="rounded bg-blue-600 px-3 py-1.5 text-xs text-white disabled:opacity-60"
                  >
                    {savingProxy ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setProxyInput(proxyPort ? String(proxyPort) : '')
                      setProxyEnabledInput(Boolean(proxyPort))
                      setProxyError('')
                      setProxyEditing(false)
                    }}
                    className="rounded border px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50"
                  >
                    {zh('取消', 'Cancel')}
                  </button>
                </div>
              )}
            </div>
            {proxyEditing && proxyError && (
              <div className="mt-1 text-sm text-red-600">{proxyError}</div>
            )}
          </section>

          <section className="rounded-lg border p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-gray-700">
                  {zh('播放器设置', 'Player Settings')}
                </h3>
                <p className="mt-1 text-xs text-gray-500">
                  {zh(
                    `当前已配置 ${Array.isArray(playerHotkeys) ? playerHotkeys.length : 0} 个快捷键`,
                    `${Array.isArray(playerHotkeys) ? playerHotkeys.length : 0} shortcuts configured`
                  )}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setPlayerSettingsOpen(true)}
                className="rounded border px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
              >
                {zh('设置快捷键', 'Configure shortcuts')}
              </button>
            </div>
          </section>

          <section className="rounded-lg border p-4">
            <h3 className="mb-4 text-sm font-semibold text-gray-700">
              {zh('目录管理', 'Directory Management')}
            </h3>
            <DirectoryManager
              open={open}
              directories={directories}
              onCreate={onCreateDirectory}
              onUpdate={onUpdateDirectory}
              onDelete={onDeleteDirectory}
            />
          </section>
        </div>
      </div>

      <PlayerSettingsModal
        open={playerSettingsOpen}
        onClose={() => setPlayerSettingsOpen(false)}
        hotkeys={playerHotkeys}
        onSave={onSavePlayerHotkeys}
      />
    </div>
  )
}
