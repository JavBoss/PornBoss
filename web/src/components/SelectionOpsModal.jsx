import { Button } from '@mui/material'
import { zh } from '@/utils/i18n'

export default function SelectionOpsModal({
  open,
  onClose,
  selectedList,
  selectedCount,
  onOpenTags,
  onOpenRemoveTags,
}) {
  if (!open) return null

  const list = Array.isArray(selectedList) ? selectedList : []
  const count = Number.isFinite(selectedCount) ? selectedCount : 0

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-lg rounded-lg bg-white p-4 shadow-xl">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-base font-semibold">{zh('已选择文件', 'Selected Files')}</h2>
          <button
            onClick={onClose}
            className="rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭', 'Close')}
          >
            ✕
          </button>
        </div>
        <div className="max-h-64 overflow-y-auto rounded border bg-gray-50 p-2 text-sm">
          {list.length === 0 ? (
            <div className="text-gray-500">{zh('暂无选择', 'No files selected')}</div>
          ) : (
            <ul className="space-y-1">
              {list.map((item) => (
                <li key={item.id} className="truncate text-gray-800">
                  {item.label}
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="outlined" size="small" onClick={onOpenTags} disabled={count === 0}>
            {zh('添加标签', 'Add Tags')}
          </Button>
          <Button
            variant="outlined"
            color="error"
            size="small"
            onClick={onOpenRemoveTags}
            disabled={count === 0}
          >
            {zh('移除标签', 'Remove Tags')}
          </Button>
          <Button variant="contained" size="small" onClick={onClose} color="primary">
            {zh('关闭', 'Close')}
          </Button>
        </div>
      </div>
    </div>
  )
}
