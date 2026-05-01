import { useEffect, useMemo, useState } from 'react'
import { fetchVideoScreenshots } from '@/api'
import { getVideoDisplayName } from '@/utils/display'
import { zh } from '@/utils/i18n'

export default function VideoScreenshotsModal({ video, onClose }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const open = Boolean(video?.id)
  const title = useMemo(() => getVideoDisplayName(video), [video])

  useEffect(() => {
    let cancelled = false
    if (!open) return undefined

    setLoading(true)
    setError('')
    setItems([])
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
          <button
            onClick={onClose}
            className="ml-3 rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭截图弹窗', 'Close screenshots modal')}
          >
            x
          </button>
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
              {items.map((item) => (
                <a
                  key={item.name}
                  href={item.url}
                  target="_blank"
                  rel="noreferrer"
                  className="group overflow-hidden rounded border border-gray-200 bg-white hover:border-gray-300"
                  title={item.name}
                >
                  <div className="aspect-video bg-gray-100">
                    <img
                      src={item.url}
                      alt={item.name}
                      loading="lazy"
                      className="h-full w-full object-contain"
                    />
                  </div>
                  <div className="truncate px-2 py-1 text-xs text-gray-600 group-hover:text-gray-900">
                    {item.name}
                  </div>
                </a>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
