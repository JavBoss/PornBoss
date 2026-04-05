import JavGrid from '@/components/JavGrid'
import Pagination from '@/components/Pagination'
import { zh } from '@/utils/i18n'

export default function JavView({
  javPage,
  javLastPage,
  javHasPrev,
  javHasNext,
  javLoading,
  javRandomMode,
  buildJavUrl,
  setJavPage,
  javItems,
  onPlay,
  onIdolClick,
  onTagClick,
  onEditTags,
  onOpenFile,
  onRevealFile,
}) {
  const contentClass = javRandomMode ? 'mt-4' : ''
  return (
    <>
      {!javRandomMode && (
        <div className="sticky-pagination mb-4 flex justify-center">
          <Pagination
            page={javPage}
            lastPage={javLastPage}
            hasPrev={javHasPrev}
            hasNext={javHasNext}
            loading={javLoading}
            buildPageUrl={({ page: targetPage }) => buildJavUrl({ page: targetPage })}
            onFirst={() => setJavPage(1)}
            onPrev={() => {
              if (javHasPrev) setJavPage(javPage - 1)
            }}
            onGoToPage={(p) => setJavPage(p)}
            onNext={() => {
              if (javHasNext) setJavPage(javPage + 1)
            }}
            onLast={() => setJavPage(javLastPage)}
          />
        </div>
      )}
      {javLoading ? (
        <div
          className={`${contentClass} flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500`}
        >
          {zh('加载中…', 'Loading...')}
        </div>
      ) : (
        <div className={contentClass}>
          <JavGrid
            items={javItems}
            onPlay={onPlay}
            onIdolClick={onIdolClick}
            onTagClick={onTagClick}
            onEditTags={onEditTags}
            onOpenFile={onOpenFile}
            onRevealFile={onRevealFile}
          />
        </div>
      )}
    </>
  )
}
