import { zh } from '@/utils/i18n'

export default function TagBar({ tags, onToggle, multiSelect, selectedIds, onSelect }) {
  return (
    <div className="flex flex-wrap gap-2">
      {tags.map((t) => {
        const checked = selectedIds?.includes(t.id)
        const count = Number.isFinite(t.count) ? t.count : null
        if (multiSelect) {
          return (
            <label
              key={t.id}
              className="inline-flex items-center gap-2 rounded px-2 py-1 text-sm text-slate-700 transition hover:bg-slate-100"
            >
              <input
                type="checkbox"
                checked={!!checked}
                onChange={() => onSelect?.(t.id)}
                className="h-3.5 w-3.5 rounded border-slate-300 text-slate-900 focus:ring-slate-400"
              />
              <span className="select-none">{t.name}</span>
              {count !== null && (
                <span className="rounded-full bg-slate-100 px-1.5 text-[10px] text-slate-500">
                  {count}
                </span>
              )}
            </label>
          )
        }
        return (
          <button
            key={t.id}
            type="button"
            className="rounded px-2 py-1 text-sm text-slate-700 transition hover:bg-slate-100"
            onClick={() => onToggle(t.name)}
            title={t.name}
            aria-label={zh(`筛选标签 ${t.name}`, `Filter tag ${t.name}`)}
          >
            <span className="inline-flex items-center gap-1">
              {t.name}
              {count !== null && (
                <span className="rounded-full bg-slate-100 px-1.5 text-[10px] text-slate-500">
                  {count}
                </span>
              )}
            </span>
          </button>
        )
      })}
    </div>
  )
}
