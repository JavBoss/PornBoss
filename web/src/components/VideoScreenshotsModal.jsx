import { useEffect, useMemo, useState } from 'react'
import CloseIcon from '@mui/icons-material/Close'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import ZoomInIcon from '@mui/icons-material/ZoomIn'
import { IconButton, Tooltip } from '@mui/material'
import { fetchVideoScreenshots } from '@/api'
import { getVideoDisplayName } from '@/utils/display'
import { zh } from '@/utils/i18n'

export default function VideoScreenshotsModal({ video, onClose, onPlayAtTime }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [previewItem, setPreviewItem] = useState(null)
  const open = Boolean(video?.id)
  const title = useMemo(() => getVideoDisplayName(video), [video])

  useEffect(() => {
    let cancelled = false
    if (!open) return undefined

    setLoading(true)
    setError('')
    setItems([])
    setPreviewItem(null)
    fetchVideoScreenshots(video.id)
      .then((nextItems) => {
        if (!cancelled) setItems(nextItems)
      })
      .catch((err) => {
        console.error(zh('加载截图失败', 'Failed to load screenshots'), err)
        if (!cancelled) setError(err?.message || zh('加载截图失败', 'Failed to load screenshots'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [open, video?.id])

  if (!open) return null

  const formatScreenshotName = (name) => {
    const stem = String(name || '')
      .replace(/\.[^.]+$/, '')
      .replace(/^mpv_/, '')
    const match = stem.match(/^(\d{2})-(\d{2})-(\d{2})(\.\d+)?$/)
    if (!match) return stem || name
    return `${match[1]}:${match[2]}:${match[3]}${match[4] || ''}`
  }

  const screenshotStartTime = (name) => {
    const stem = String(name || '')
      .replace(/\.[^.]+$/, '')
      .replace(/^mpv_/, '')
    const match = stem.match(/^(\d{2})-(\d{2})-(\d{2})(\.\d+)?$/)
    if (!match) return null
    return (
      Number.parseInt(match[1], 10) * 3600 +
      Number.parseInt(match[2], 10) * 60 +
      Number.parseInt(match[3], 10) +
      Number.parseFloat(match[4] || '0')
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4 py-6">
      <div className="flex max-h-full w-full max-w-5xl flex-col rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold text-gray-900">
              {zh('视频截图', 'Video Screenshots')}
            </h2>
            <div className="truncate text-xs text-gray-500" title={title}>
              {title}
            </div>
          </div>
          <IconButton
            size="small"
            onClick={onClose}
            aria-label={zh('关闭截图弹窗', 'Close screenshots modal')}
          >
            <CloseIcon fontSize="inherit" />
          </IconButton>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          {loading ? (
            <div className="flex min-h-48 items-center justify-center rounded border border-dashed border-gray-200 text-sm text-gray-500">
              {zh('加载中...', 'Loading...')}
            </div>
          ) : error ? (
            <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          ) : items.length === 0 ? (
            <div className="flex min-h-48 items-center justify-center rounded border border-dashed border-gray-200 text-sm text-gray-500">
              {zh('暂无截图', 'No screenshots')}
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {items.map((item) => {
                const displayName = formatScreenshotName(item.name)
                const startTime = screenshotStartTime(item.name)
                return (
                  <div
                    key={item.name}
                    className="group overflow-hidden rounded border border-gray-200 bg-white hover:border-gray-300"
                    title={displayName}
                  >
                    <div className="relative aspect-video bg-gray-100">
                      <img
                        src={item.url}
                        alt={item.name}
                        loading="lazy"
                        className="h-full w-full object-contain"
                      />
                      <div className="absolute inset-0 flex items-center justify-center gap-2 bg-black/0 opacity-0 transition group-hover:bg-black/35 group-hover:opacity-100">
                        <Tooltip title={zh('放大图片', 'Enlarge image')}>
                          <IconButton
                            size="small"
                            onClick={() => setPreviewItem(item)}
                            aria-label={zh('放大图片', 'Enlarge image')}
                            className="!bg-white/90 !text-gray-900 hover:!bg-white"
                          >
                            <ZoomInIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title={zh('从此处播放', 'Play from here')}>
                          <span>
                            <IconButton
                              size="small"
                              onClick={() => onPlayAtTime?.(video, startTime)}
                              disabled={startTime == null}
                              aria-label={zh('从此处播放', 'Play from here')}
                              className="!bg-white/90 !text-gray-900 hover:!bg-white disabled:!opacity-50"
                            >
                              <PlayArrowIcon fontSize="small" />
                            </IconButton>
                          </span>
                        </Tooltip>
                      </div>
                    </div>
                    <div className="truncate px-2 py-1 text-xs text-gray-600 group-hover:text-gray-900">
                      {displayName}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
      {previewItem ? (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/85 p-4">
          <img
            src={previewItem.url}
            alt={previewItem.name}
            className="max-h-full max-w-full object-contain"
          />
          <IconButton
            size="small"
            onClick={() => setPreviewItem(null)}
            aria-label={zh('关闭大图', 'Close enlarged image')}
            className="!absolute !right-4 !top-4 !bg-white/90 !text-gray-900 hover:!bg-white"
          >
            <CloseIcon fontSize="small" />
          </IconButton>
        </div>
      ) : null}
    </div>
  )
}
