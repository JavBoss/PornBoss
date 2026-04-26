import { useEffect, useState } from 'react'

import DirectoryManager from '@/components/DirectoryManager'
import PlayerSettingsModal from '@/components/PlayerSettingsModal'
import { parsePlayerHotkeys } from '@/utils/playerHotkeys'
import { zh } from '@/utils/i18n'

const SETTINGS_SECTIONS = [
  {
    id: 'directories',
    title: { zh: '目录管理', en: 'Directory Management' },
    summary: { zh: '管理扫描目录与路径', en: 'Manage watched folders and paths' },
  },
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
]

const PLAYER_BASIC_DEFAULTS = {
  windowWidth: 70,
  windowHeight: 70,
  windowUseAutofit: false,
  ontop: true,
  volume: 70,
}

export default function GlobalSettingsModal({
  open,
  onClose,
  directories,
  onCreateDirectory,
  onUpdateDirectory,
  onDeleteDirectory,
  proxyPort,
  onSaveProxyPort,
  playerWindowWidth,
  playerWindowHeight,
  playerWindowUseAutofit,
  playerOntop,
  playerVolume,
  onSavePlayerBasicSettings,
  playerHotkeys,
  onSavePlayerHotkeys,
}) {
  const [proxyInput, setProxyInput] = useState('')
  const [proxyError, setProxyError] = useState('')
  const [savingProxy, setSavingProxy] = useState(false)
  const [proxyEditing, setProxyEditing] = useState(false)
  const [proxyEnabledInput, setProxyEnabledInput] = useState(false)
  const [activeSection, setActiveSection] = useState('directories')
  const [playerTab, setPlayerTab] = useState('basic')
  const [playerBasicError, setPlayerBasicError] = useState('')
  const [playerBasicSuccess, setPlayerBasicSuccess] = useState('')
  const [savingPlayerBasic, setSavingPlayerBasic] = useState(false)
  const [playerWindowWidthInput, setPlayerWindowWidthInput] = useState('')
  const [playerWindowHeightInput, setPlayerWindowHeightInput] = useState('')
  const [playerWindowUseAutofitInput, setPlayerWindowUseAutofitInput] = useState(false)
  const [playerOntopInput, setPlayerOntopInput] = useState(true)
  const [playerVolumeInput, setPlayerVolumeInput] = useState('')

  const normalizedPlayerHotkeys = parsePlayerHotkeys(playerHotkeys)

  const resetPlayerBasicInputs = () => {
    setPlayerWindowWidthInput(String(PLAYER_BASIC_DEFAULTS.windowWidth))
    setPlayerWindowHeightInput(String(PLAYER_BASIC_DEFAULTS.windowHeight))
    setPlayerWindowUseAutofitInput(PLAYER_BASIC_DEFAULTS.windowUseAutofit)
    setPlayerOntopInput(PLAYER_BASIC_DEFAULTS.ontop)
    setPlayerVolumeInput(String(PLAYER_BASIC_DEFAULTS.volume))
    setPlayerBasicError('')
    setPlayerBasicSuccess('')
  }

  useEffect(() => {
    if (open) {
      setProxyInput(proxyPort ? String(proxyPort) : '')
      setProxyEnabledInput(Boolean(proxyPort))
      setProxyEditing(false)
      setProxyError('')
      setPlayerTab('basic')
      setPlayerBasicError('')
      setPlayerBasicSuccess('')
      setPlayerWindowWidthInput(String(playerWindowWidth ?? PLAYER_BASIC_DEFAULTS.windowWidth))
      setPlayerWindowHeightInput(String(playerWindowHeight ?? PLAYER_BASIC_DEFAULTS.windowHeight))
      setPlayerWindowUseAutofitInput(
        playerWindowUseAutofit ?? PLAYER_BASIC_DEFAULTS.windowUseAutofit
      )
      setPlayerOntopInput(playerOntop ?? PLAYER_BASIC_DEFAULTS.ontop)
      setPlayerVolumeInput(String(playerVolume ?? PLAYER_BASIC_DEFAULTS.volume))
    }
  }, [
    open,
    proxyPort,
    playerWindowWidth,
    playerWindowHeight,
    playerWindowUseAutofit,
    playerOntop,
    playerVolume,
  ])

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
        <div className="mb-4 flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setPlayerTab('basic')}
            className={`rounded-xl px-3 py-1.5 text-sm ${
              playerTab === 'basic'
                ? 'bg-zinc-900 text-white'
                : 'border border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-50'
            }`}
          >
            {zh('基础设置', 'Basic Settings')}
          </button>
          <button
            type="button"
            onClick={() => setPlayerTab('hotkeys')}
            className={`rounded-xl px-3 py-1.5 text-sm ${
              playerTab === 'hotkeys'
                ? 'bg-zinc-900 text-white'
                : 'border border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-50'
            }`}
          >
            {zh('快捷键设置', 'Shortcut Settings')}
          </button>
        </div>

        {playerTab === 'basic' ? (
          <div>
            <div className="space-y-6">
              <section className="space-y-3">
                <h4 className="text-sm font-semibold text-zinc-800">
                  {zh('初始窗口大小', 'Initial Window Size')}
                </h4>
                <div className="flex flex-col gap-3">
                  <div className="grid gap-3 md:max-w-xl md:grid-cols-2">
                    <label className="flex items-center gap-2 text-xs font-medium text-zinc-500">
                      <span className="shrink-0">{zh('宽度', 'Width')}</span>
                      <div className="flex min-w-0 flex-1 items-center gap-2">
                        <input
                          value={playerWindowWidthInput}
                          onChange={(e) => {
                            setPlayerWindowWidthInput(e.target.value)
                            setPlayerBasicError('')
                            setPlayerBasicSuccess('')
                          }}
                          inputMode="numeric"
                          className="w-full min-w-0 rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-800"
                        />
                        <span className="text-sm text-zinc-500">%</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 text-xs font-medium text-zinc-500">
                      <span className="shrink-0">{zh('高度', 'Height')}</span>
                      <div className="flex min-w-0 flex-1 items-center gap-2">
                        <input
                          value={playerWindowHeightInput}
                          onChange={(e) => {
                            setPlayerWindowHeightInput(e.target.value)
                            setPlayerBasicError('')
                            setPlayerBasicSuccess('')
                          }}
                          inputMode="numeric"
                          className="w-full min-w-0 rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-800"
                        />
                        <span className="text-sm text-zinc-500">%</span>
                      </div>
                    </label>
                  </div>

                  <div className="flex flex-wrap items-center gap-3">
                    <label className="flex items-center gap-3 text-sm font-medium text-zinc-600">
                      <input
                        type="checkbox"
                        checked={playerWindowUseAutofitInput}
                        onChange={(e) => {
                          setPlayerWindowUseAutofitInput(e.target.checked)
                          setPlayerBasicError('')
                          setPlayerBasicSuccess('')
                        }}
                        className="h-4 w-4 rounded"
                      />
                      <span>{zh('自动调节窗口大小', 'Automatically adjust window size')}</span>
                    </label>
                    <span className="text-xs text-zinc-500">
                      {playerWindowUseAutofitInput
                        ? zh(
                            '开启后按最大宽高限制窗口，并保持视频纵横比。',
                            'When enabled, the window is limited by max width and height while preserving aspect ratio.'
                          )
                        : zh(
                            '默认关闭，强制使用指定宽高。',
                            'Disabled by default and forces the specified width and height.'
                          )}
                    </span>
                  </div>
                </div>
              </section>

              <section className="space-y-3 border-t border-zinc-200 pt-5">
                <div className="flex flex-wrap items-center gap-3">
                  <h4 className="text-sm font-semibold text-zinc-800">
                    {zh('初始音量', 'Initial Volume')}
                  </h4>
                  <div className="flex w-full items-center gap-2 md:max-w-sm">
                    <input
                      value={playerVolumeInput}
                      onChange={(e) => {
                        setPlayerVolumeInput(e.target.value)
                        setPlayerBasicError('')
                        setPlayerBasicSuccess('')
                      }}
                      inputMode="numeric"
                      className="w-full rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-800"
                    />
                    <span className="text-sm text-zinc-500">%</span>
                  </div>
                </div>
                <p className="text-xs text-zinc-500">
                  {zh(
                    '控制 mpv 启动时的默认音量，范围 0-130。',
                    'Controls the default mpv startup volume, range 0-130.'
                  )}
                </p>
              </section>

              <section className="space-y-3 border-t border-zinc-200 pt-5">
                <label className="flex items-center gap-3 text-sm font-semibold text-zinc-800">
                  <input
                    type="checkbox"
                    checked={playerOntopInput}
                    onChange={(e) => {
                      setPlayerOntopInput(e.target.checked)
                      setPlayerBasicError('')
                      setPlayerBasicSuccess('')
                    }}
                    className="h-4 w-4 rounded"
                  />
                  <span>{zh('播放器强行置顶', 'Keep Player On Top')}</span>
                </label>
                <p className="text-xs text-zinc-500">
                  {zh(
                    '默认开启，使 mpv 播放器窗口保持置顶。',
                    'Enabled by default to keep the mpv player window on top.'
                  )}
                </p>
              </section>
            </div>

            {playerBasicError && (
              <div className="mt-3 text-sm text-red-600">{playerBasicError}</div>
            )}
            {playerBasicSuccess && (
              <div className="mt-3 text-sm text-emerald-600">{playerBasicSuccess}</div>
            )}

            <div className="mt-4 flex justify-end gap-2">
              <button
                type="button"
                onClick={resetPlayerBasicInputs}
                disabled={savingPlayerBasic}
                className="rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-sm text-zinc-700 hover:bg-zinc-50 disabled:opacity-60"
              >
                {zh('恢复默认', 'Restore Defaults')}
              </button>
              <button
                type="button"
                onClick={async () => {
                  setPlayerBasicError('')
                  setPlayerBasicSuccess('')
                  const width = Number.parseInt(playerWindowWidthInput, 10)
                  const height = Number.parseInt(playerWindowHeightInput, 10)
                  const volume = Number.parseInt(playerVolumeInput, 10)
                  if (!Number.isFinite(width) || width < 10 || width > 100) {
                    setPlayerBasicError(
                      zh('初始宽度请输入 10-100', 'Initial width must be between 10 and 100')
                    )
                    return
                  }
                  if (!Number.isFinite(height) || height < 10 || height > 100) {
                    setPlayerBasicError(
                      zh('初始高度请输入 10-100', 'Initial height must be between 10 and 100')
                    )
                    return
                  }
                  if (!Number.isFinite(volume) || volume < 0 || volume > 130) {
                    setPlayerBasicError(
                      zh('初始音量请输入 0-130', 'Initial volume must be between 0 and 130')
                    )
                    return
                  }

                  setSavingPlayerBasic(true)
                  try {
                    await onSavePlayerBasicSettings?.({
                      player_window_width: width,
                      player_window_height: height,
                      player_window_use_autofit: playerWindowUseAutofitInput,
                      player_ontop: playerOntopInput,
                      player_volume: volume,
                    })
                    setPlayerBasicSuccess(zh('基础设置保存成功', 'Basic settings saved'))
                  } catch (err) {
                    setPlayerBasicError(err.message || zh('保存失败', 'Save failed'))
                  } finally {
                    setSavingPlayerBasic(false)
                  }
                }}
                disabled={savingPlayerBasic}
                className="rounded-xl bg-blue-600 px-3 py-1.5 text-sm text-white disabled:opacity-60"
              >
                {savingPlayerBasic ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
              </button>
            </div>
          </div>
        ) : (
          <>
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-zinc-800">
                {zh('快捷键设置', 'Shortcut Settings')}
              </h4>
              <p className="mt-1 text-xs text-zinc-500">
                {zh(
                  '正数表示增加，负数表示减少。`Space` 和 `Escape` 仍固定用于播放/暂停和关闭播放器。',
                  'Positive numbers increase, negative numbers decrease. `Space` and `Escape` remain reserved for play/pause and close.'
                )}
              </p>
            </div>
            <PlayerSettingsModal hotkeys={normalizedPlayerHotkeys} onSave={onSavePlayerHotkeys} />
          </>
        )}
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
                      ? ''
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
                      </div>
                      {badgeText ? (
                        <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-600">
                          {badgeText}
                        </span>
                      ) : null}
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
