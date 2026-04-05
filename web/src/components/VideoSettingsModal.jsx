import { zh } from '@/utils/i18n'

export default function VideoSettingsModal({
  open,
  onClose,
  pageSizeInput,
  onPageSizeChange,
  sortInput,
  onSortChange,
  onSave,
}) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-sm rounded-lg bg-white p-3 shadow-xl">
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-base font-semibold">{zh('视频设置', 'Video Settings')}</h2>
          <button
            onClick={onClose}
            className="rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭设置', 'Close settings')}
          >
            ✕
          </button>
        </div>
        <div className="space-y-2">
          <label className="flex items-center justify-between gap-3 text-sm font-medium text-gray-700">
            <span>{zh('每页视频数量', 'Videos per page')}</span>
            <input
              type="number"
              min="1"
              value={pageSizeInput}
              onChange={(e) => onPageSizeChange?.(e.target.value)}
              className="w-24 rounded border px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <div className="text-sm font-medium text-gray-700">{zh('分页排序', 'Sort order')}</div>
          <label
            htmlFor="sort-recent"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按最近加入排序', 'Sort by recently added')}
          >
            <input
              id="sort-recent"
              type="radio"
              name="sort"
              value="recent"
              checked={sortInput === 'recent'}
              onChange={() => onSortChange?.('recent')}
            />
            <div className="text-sm font-semibold">
              {zh('最近加入（默认）', 'Recently added (default)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按创建时间倒序', 'Newest first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="sort-filename"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按文件名排序', 'Sort by filename')}
          >
            <input
              id="sort-filename"
              type="radio"
              name="sort"
              value="filename"
              checked={sortInput === 'filename'}
              onChange={() => onSortChange?.('filename')}
            />
            <div className="text-sm font-semibold">
              {zh('文件名', 'Filename')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按文件名排序', 'Alphabetical order')}
              </span>
            </div>
          </label>
          <label
            htmlFor="sort-duration"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按视频时长排序', 'Sort by duration')}
          >
            <input
              id="sort-duration"
              type="radio"
              name="sort"
              value="duration"
              checked={sortInput === 'duration'}
              onChange={() => onSortChange?.('duration')}
            />
            <div className="text-sm font-semibold">
              {zh('时长（长→短）', 'Duration (long to short)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按视频时长降序', 'Longest first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="sort-play-count"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按播放次数排序', 'Sort by play count')}
          >
            <input
              id="sort-play-count"
              type="radio"
              name="sort"
              value="play_count"
              checked={sortInput === 'play_count'}
              onChange={() => onSortChange?.('play_count')}
            />
            <div className="text-sm font-semibold">
              {zh('播放次数（多→少）', 'Play count (high to low)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按播放次数降序', 'Most played first')}
              </span>
            </div>
          </label>
        </div>
        <div className="mt-3 flex justify-end">
          <button onClick={onClose} className="rounded border px-3 py-1 text-sm hover:bg-gray-50">
            {zh('取消', 'Cancel')}
          </button>
          <button
            onClick={onSave}
            className="ml-2 rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700"
          >
            {zh('保存', 'Save')}
          </button>
        </div>
      </div>
    </div>
  )
}
