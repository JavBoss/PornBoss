import { Button } from '@mui/material'
import Pagination from './Pagination'
import VideoGrid from './VideoGrid'

export default function VideoView({
  selectedCount,
  clearSelection,
  setSelectionOpsOpen,
  page,
  lastPage,
  canPrev,
  canNext,
  loading,
  randomMode,
  buildVideoUrl,
  setPage,
  goToLastPage,
  videos,
  selectedVideoIds,
  toggleSelectVideo,
  onToggleSelectPage,
  openPlayer,
  setTagPickerFor,
  onTagClick,
}) {
  const pageIds = videos.map((video) => video?.id).filter(Boolean)
  const pageSelectable = pageIds.length > 0
  const pageAllSelected = pageSelectable && pageIds.every((id) => selectedVideoIds.has(id))
  const hasSelection = selectedCount > 0

  return (
    <>
      <div className="sticky-pagination mb-4 grid grid-cols-[auto_1fr_auto] items-center gap-3">
        <div className="flex items-center gap-1 overflow-x-auto overflow-y-hidden whitespace-nowrap">
          {hasSelection && (
            <>
              <span className="rounded-full bg-sky-50 px-2 py-0.5 text-xs font-medium leading-5 text-sky-700">
                已选 {selectedCount} 项
              </span>
              <Button
                variant="outlined"
                size="small"
                onClick={onToggleSelectPage}
                disabled={!pageSelectable}
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                {pageAllSelected ? '取消本页' : '全选本页'}
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => setSelectionOpsOpen(true)}
                disabled={selectedCount === 0}
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                操作
              </Button>
              <Button
                variant="text"
                size="small"
                onClick={clearSelection}
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                清空
              </Button>
            </>
          )}
        </div>
        <div className="flex justify-center">
          {!randomMode && (
            <Pagination
              page={page}
              lastPage={lastPage}
              hasPrev={canPrev}
              hasNext={canNext}
              loading={loading}
              buildPageUrl={({ page: targetPage }) =>
                buildVideoUrl({ page: targetPage, random: false })
              }
              onFirst={() => setPage(1)}
              onPrev={() => {
                if (canPrev) setPage(page - 1)
              }}
              onGoToPage={(p) => setPage(p)}
              onNext={() => {
                if (canNext) setPage(page + 1)
              }}
              onLast={() => {
                goToLastPage()
              }}
            />
          )}
        </div>
        <div className="pointer-events-none flex items-center justify-end gap-1 overflow-hidden whitespace-nowrap opacity-0">
          {hasSelection && (
            <>
              <span className="rounded-full px-2 py-0.5 text-xs font-medium leading-5">
                已选 {selectedCount} 项
              </span>
              <Button
                variant="outlined"
                size="small"
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                {pageAllSelected ? '取消本页' : '全选本页'}
              </Button>
              <Button
                variant="outlined"
                size="small"
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                操作
              </Button>
              <Button
                variant="text"
                size="small"
                className="!min-h-0 !min-w-0 !px-2 !py-0.5 !text-xs !leading-5"
              >
                清空
              </Button>
            </>
          )}
        </div>
      </div>
      {loading ? (
        <div className="mt-4 flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
          加载中…
        </div>
      ) : (
        <VideoGrid
          videos={videos}
          selectedIds={selectedVideoIds}
          onToggleSelect={toggleSelectVideo}
          onPlay={(video) => openPlayer(video)}
          onOpenTagPicker={(vid) => setTagPickerFor(vid)}
          onTagClick={onTagClick}
        />
      )}
    </>
  )
}
