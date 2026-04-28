import SwapVertIcon from '@mui/icons-material/SwapVert'
import {
  IDOL_SORT_OPTIONS,
  JAV_SORT_OPTIONS,
  findSortOption,
  reverseSortValue,
  sortLabelParts,
} from '@/constants/jav'
import { zh } from '@/utils/i18n'

function SortText({ option, value }) {
  const parts = sortLabelParts(option, value, zh)

  return (
    <span className="truncate text-sm font-semibold">
      <span>{parts.label}</span>
      <span className="font-normal text-gray-500">{parts.separator}</span>
      <span className="font-normal text-gray-500">{parts.direction}</span>
    </span>
  )
}

function SortOptionRow({ option, name, inputValue, onChange }) {
  const active = findSortOption([option], inputValue)
  const displayValue = active ? inputValue : option.defaultValue
  const id = `${name}-${option.base}`

  return (
    <div className="flex items-center gap-2 rounded border bg-white px-3 py-1.5 hover:border-blue-500">
      <label htmlFor={id} className="flex min-w-0 flex-1 cursor-pointer items-center gap-3">
        <input
          id={id}
          type="radio"
          name={name}
          value={displayValue}
          checked={Boolean(active)}
          onChange={() => onChange?.(displayValue)}
        />
        <SortText option={option} value={displayValue} />
      </label>
      <button
        type="button"
        onClick={() => onChange?.(reverseSortValue([option], displayValue, option.defaultValue))}
        className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded border border-gray-200 text-gray-500 hover:border-blue-400 hover:bg-blue-50 hover:text-blue-700"
        title={zh('反转排序', 'Reverse sort')}
        aria-label={zh(`反转${option.label[0]}排序`, `Reverse ${option.label[1]} sort`)}
      >
        <SwapVertIcon fontSize="inherit" />
      </button>
    </div>
  )
}

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
      <div className="w-full max-w-2xl rounded-lg bg-white p-4 shadow-xl">
        <div className="mb-2 flex items-center justify-between">
          <div />
          <button
            onClick={onClose}
            className="rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭设置', 'Close settings')}
          >
            ✕
          </button>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <section className="space-y-3 rounded-lg border border-gray-200 bg-gray-50/60 p-3">
            <div className="border-b border-gray-200 pb-2 text-sm font-semibold text-gray-800">
              {zh('Jav设置', 'Jav Settings')}
            </div>
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
            <div className="text-sm font-medium text-gray-700">
              {zh('默认排序', 'Default sort')}
            </div>
            {JAV_SORT_OPTIONS.map((option) => (
              <SortOptionRow
                key={option.base}
                option={option}
                name="jav-sort"
                inputValue={javSortInput}
                onChange={onJavSortChange}
              />
            ))}
          </section>

          <section className="space-y-3 rounded-lg border border-gray-200 bg-gray-50/60 p-3">
            <div className="border-b border-gray-200 pb-2 text-sm font-semibold text-gray-800">
              {zh('女优设置', 'Idol Settings')}
            </div>
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
            <div className="text-sm font-medium text-gray-700">
              {zh('女优排序', 'Idol sorting')}
            </div>
            {IDOL_SORT_OPTIONS.map((option) => (
              <SortOptionRow
                key={option.base}
                option={option}
                name="idol-sort"
                inputValue={idolSortInput}
                onChange={onIdolSortChange}
              />
            ))}
          </section>
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
