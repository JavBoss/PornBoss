import VideoCard from '@/components/VideoCard'

export default function VideoGrid({
  videos,
  selectedIds,
  onToggleSelect,
  onPlay,
  onOpenFile,
  openFileLabel,
  onOpenTagPicker,
  onTagClick,
}) {
  return (
    <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      {videos.map((v) => (
        <VideoCard
          key={v.id}
          video={v}
          checked={selectedIds.has(v.id)}
          onToggle={() => onToggleSelect(v)}
          onPlay={onPlay}
          onOpenFile={onOpenFile}
          openFileLabel={openFileLabel}
          onOpenTagPicker={() => onOpenTagPicker(v.id)}
          onTagClick={onTagClick}
        />
      ))}
    </div>
  )
}
