import { zh } from '@/utils/i18n'

export default function JavSettingsModal({
  open,
  onClose,
  javPageSizeInput,
  onJavPageSizeChange,
  idolPageSizeInput,
  onIdolPageSizeChange,
  javSortInput,
  onJavSortChange,
  idolSortInput,
  onIdolSortChange,
  onSave,
}) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-sm rounded-lg bg-white p-3 shadow-xl">
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-base font-semibold">{zh('JAV 设置', 'JAV Settings')}</h2>
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
            <span>{zh('每页 JAV 数量', 'JAVs per page')}</span>
            <input
              type="number"
              min="1"
              value={javPageSizeInput}
              onChange={(e) => onJavPageSizeChange?.(e.target.value)}
              className="w-24 rounded border px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <label className="flex items-center justify-between gap-3 text-sm font-medium text-gray-700">
            <span>{zh('每页 女优 数量', 'Idols per page')}</span>
            <input
              type="number"
              min="1"
              value={idolPageSizeInput}
              onChange={(e) => onIdolPageSizeChange?.(e.target.value)}
              className="w-24 rounded border px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <div className="text-sm font-medium text-gray-700">{zh('作品排序', 'Title sorting')}</div>
          <label
            htmlFor="jav-sort-recent"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按加入时间排序', 'Sort by import time')}
          >
            <input
              id="jav-sort-recent"
              type="radio"
              name="jav-sort"
              value="recent"
              checked={javSortInput === 'recent'}
              onChange={() => onJavSortChange?.('recent')}
            />
            <div className="text-sm font-semibold">
              {zh('加入时间（默认）', 'Added time (default)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('最近入库的 JAV 优先', 'Recently imported first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-code"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按番号排序', 'Sort by code')}
          >
            <input
              id="jav-sort-code"
              type="radio"
              name="jav-sort"
              value="code"
              checked={javSortInput === 'code'}
              onChange={() => onJavSortChange?.('code')}
            />
            <div className="text-sm font-semibold">
              {zh('番号', 'Code')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按番号字母顺序', 'Alphabetical by code')}
              </span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-release"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按发行时间排序', 'Sort by release date')}
          >
            <input
              id="jav-sort-release"
              type="radio"
              name="jav-sort"
              value="release"
              checked={javSortInput === 'release'}
              onChange={() => onJavSortChange?.('release')}
            />
            <div className="text-sm font-semibold">
              {zh('发行时间', 'Release date')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('最新发行优先', 'Newest releases first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-play-count"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按播放次数排序', 'Sort by play count')}
          >
            <input
              id="jav-sort-play-count"
              type="radio"
              name="jav-sort"
              value="play_count"
              checked={javSortInput === 'play_count'}
              onChange={() => onJavSortChange?.('play_count')}
            />
            <div className="text-sm font-semibold">
              {zh('播放次数（多→少）', 'Play count (high to low)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('按作品累计播放次数', 'By total plays across files')}
              </span>
            </div>
          </label>
          <div className="text-sm font-medium text-gray-700">{zh('女优排序', 'Idol sorting')}</div>
          <label
            htmlFor="idol-sort-work"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按作品数量排序', 'Sort by work count')}
          >
            <input
              id="idol-sort-work"
              type="radio"
              name="idol-sort"
              value="work"
              checked={idolSortInput === 'work'}
              onChange={() => onIdolSortChange?.('work')}
            />
            <div className="text-sm font-semibold">
              {zh('作品数量（默认）', 'Work count (default)')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('作品越多越靠前', 'More works first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-birth"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按出生日期排序', 'Sort by birth date')}
          >
            <input
              id="idol-sort-birth"
              type="radio"
              name="idol-sort"
              value="birth"
              checked={idolSortInput === 'birth'}
              onChange={() => onIdolSortChange?.('birth')}
            />
            <div className="text-sm font-semibold">
              {zh('年龄', 'Age')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由小到大（更年轻优先）', 'Younger first')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-height"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按身高排序', 'Sort by height')}
          >
            <input
              id="idol-sort-height"
              type="radio"
              name="idol-sort"
              value="height"
              checked={idolSortInput === 'height'}
              onChange={() => onIdolSortChange?.('height')}
            />
            <div className="text-sm font-semibold">
              {zh('身高', 'Height')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由低到高', 'Low to high')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-bust"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按胸围排序', 'Sort by bust')}
          >
            <input
              id="idol-sort-bust"
              type="radio"
              name="idol-sort"
              value="bust"
              checked={idolSortInput === 'bust'}
              onChange={() => onIdolSortChange?.('bust')}
            />
            <div className="text-sm font-semibold">
              {zh('胸围', 'Bust')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由大到小', 'High to low')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-hips"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按臀围排序', 'Sort by hips')}
          >
            <input
              id="idol-sort-hips"
              type="radio"
              name="idol-sort"
              value="hips"
              checked={idolSortInput === 'hips'}
              onChange={() => onIdolSortChange?.('hips')}
            />
            <div className="text-sm font-semibold">
              {zh('臀围', 'Hips')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由大到小', 'High to low')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-waist"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按腰围排序', 'Sort by waist')}
          >
            <input
              id="idol-sort-waist"
              type="radio"
              name="idol-sort"
              value="waist"
              checked={idolSortInput === 'waist'}
              onChange={() => onIdolSortChange?.('waist')}
            />
            <div className="text-sm font-semibold">
              {zh('腰围', 'Waist')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由小到大', 'Low to high')}
              </span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-cup"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label={zh('按罩杯排序', 'Sort by cup')}
          >
            <input
              id="idol-sort-cup"
              type="radio"
              name="idol-sort"
              value="cup"
              checked={idolSortInput === 'cup'}
              onChange={() => onIdolSortChange?.('cup')}
            />
            <div className="text-sm font-semibold">
              {zh('罩杯', 'Cup')}
              <span className="ml-2 text-xs font-normal text-gray-500">
                {zh('由大到小', 'High to low')}
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
