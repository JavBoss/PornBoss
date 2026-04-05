import { useMemo } from 'react'
import { IconButton, Tooltip } from '@mui/material'
import Fade from '@mui/material/Fade'
import EditIcon from '@mui/icons-material/Edit'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import FolderOpenIcon from '@mui/icons-material/FolderOpen'
import LanguageIcon from '@mui/icons-material/Language'

import { isUserJavTag } from '@/constants/jav'
import { zh } from '@/utils/i18n'

export default function JavGrid({
  items,
  onPlay,
  onIdolClick,
  onTagClick,
  onEditTags,
  onOpenFile,
  onRevealFile,
}) {
  const hasItems = Array.isArray(items) && items.length > 0
  if (!hasItems) {
    return (
      <div className="mt-4 flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
        {zh('暂无 JAV 数据', 'No JAV data')}
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {items.map((item) => (
        <JavCard
          key={item.id || item.code}
          item={item}
          onPlay={onPlay}
          onIdolClick={onIdolClick}
          onTagClick={onTagClick}
          onEditTags={onEditTags}
          onOpenFile={onOpenFile}
          onRevealFile={onRevealFile}
        />
      ))}
    </div>
  )
}

function JavCard({ item, onPlay, onIdolClick, onTagClick, onEditTags, onOpenFile, onRevealFile }) {
  const primaryVideo = useMemo(() => (item?.videos || [])[0], [item])
  const cover = item?.code ? `/jav/${encodeURIComponent(item.code)}/cover` : null

  const release =
    item?.release_unix && Number.isFinite(item.release_unix)
      ? new Date(item.release_unix * 1000)
      : null
  const releaseText = release ? release.toISOString().slice(0, 10) : zh('未知', 'Unknown')
  const durationText = item?.duration_min
    ? zh(`${item.duration_min} 分钟`, `${item.duration_min} min`)
    : ''
  const titleText = [item?.code, item?.title || item?.code || zh('未知标题', 'Untitled')]
    .filter(Boolean)
    .join(' ')
  const videos = item?.videos || []
  const openableVideos = videos.filter((video) =>
    Boolean(video?.path && (video?.directory?.path || video?.directory_path))
  )
  const canOpen = openableVideos.length > 0
  const code = item?.code?.trim()
  const encodedCode = code ? encodeURIComponent(code) : ''
  const externalLinks = encodedCode
    ? [
        {
          key: 'javlibrary',
          name: 'JavLibrary',
          href: `https://www.javlibrary.com/cn/vl_searchbyid.php?keyword=${encodedCode}`,
          icon: '/ico/javlibrary.ico',
        },
        {
          key: 'javbus',
          name: 'JavBus',
          href: `https://www.javbus.com/${encodedCode}`,
          icon: '/ico/javbus.ico',
        },
      ]
    : []

  const handleOpenFile = (event) => {
    event.stopPropagation()
    if (!canOpen) return
    onOpenFile?.(openableVideos[0] || primaryVideo, item)
  }

  const handleRevealFile = (event) => {
    event.stopPropagation()
    if (!canOpen) return
    onRevealFile?.(openableVideos[0] || primaryVideo, item)
  }

  const canPlay = Boolean(primaryVideo && primaryVideo.id)
  const handlePlay = (event) => {
    event?.stopPropagation()
    if (!canPlay) return
    onPlay?.(primaryVideo, item)
  }
  const handleEditTags = (event) => {
    event?.stopPropagation()
    onEditTags?.(item)
  }
  const tags = Array.isArray(item?.tags) ? item.tags : []
  const showEditTags = typeof onEditTags === 'function'

  return (
    <div className="flex flex-col overflow-hidden rounded-lg border bg-white shadow-sm transition hover:shadow-lg">
      <div className="group relative h-64 bg-gray-100">
        {cover ? (
          <img src={cover} alt={item?.code} className="h-full w-full object-cover" loading="lazy" />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-gray-100 to-gray-200 text-lg font-semibold text-gray-600">
            {item?.code || zh('未知番号', 'Unknown code')}
          </div>
        )}
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/0 text-white opacity-0 transition-opacity group-hover:opacity-100">
          <button
            onClick={handlePlay}
            disabled={!canPlay}
            className={`pointer-events-auto rounded-full p-3 ${
              canPlay ? 'bg-black/60 hover:bg-black/80' : 'cursor-not-allowed bg-black/30'
            }`}
            aria-label={zh('播放', 'Play')}
            title={zh('播放', 'Play')}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="currentColor"
              className="h-10 w-10"
            >
              <path d="M8 5v14l11-7z" />
            </svg>
          </button>
        </div>
      </div>
      <div className="flex flex-1 flex-col gap-2 p-3">
        <div className="line-clamp-2 text-sm font-semibold leading-tight" title={titleText}>
          {titleText}
        </div>
        <div className="text-xs text-gray-600">
          {durationText || zh('时长未知', 'Unknown duration')}
          {releaseText ? ` · ${releaseText}` : ''}
        </div>
        {Array.isArray(item?.idols) && item.idols.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {item.idols.map((idol) => (
              <button
                key={idol.id || idol.name}
                type="button"
                className="rounded-full bg-purple-100 px-2 py-1 text-xs font-medium text-purple-700 transition hover:bg-purple-200"
                onClick={() => onIdolClick?.(idol.name)}
              >
                {idol.name}
              </button>
            ))}
          </div>
        )}
        {tags.length > 0 && (
          <div className="flex flex-wrap items-center gap-1">
            {tags.map((tag) => {
              const isUser = isUserJavTag(tag)
              const tagClass = isUser
                ? 'bg-emerald-500 hover:bg-emerald-600'
                : 'bg-orange-500 hover:bg-orange-600'
              return (
                <button
                  key={tag.id || tag.name}
                  type="button"
                  className={`rounded-full px-2 py-1 text-xs font-medium text-white transition ${tagClass}`}
                  onClick={() => onTagClick?.(tag)}
                >
                  {tag.name}
                </button>
              )
            })}
          </div>
        )}
        <div className="flex flex-wrap items-center gap-2">
          {Array.isArray(item?.videos) && item.videos.length > 1 && (
            <span className="text-xs text-gray-500">
              {zh(`共 ${item.videos.length} 个视频`, `${item.videos.length} video files`)}
            </span>
          )}
          {externalLinks.length > 0 && (
            <div className="group relative flex items-center">
              <IconButton
                size="small"
                aria-label={zh('外部站点', 'External links')}
                className="h-6 w-6"
                onClick={(event) => event.stopPropagation()}
              >
                <LanguageIcon fontSize="inherit" />
              </IconButton>
              <div className="pointer-events-none absolute bottom-full left-0 z-20 flex items-center gap-2 rounded-full border border-gray-200 bg-white/95 px-2 py-1 text-xs text-gray-700 opacity-0 shadow-lg backdrop-blur transition group-hover:pointer-events-auto group-hover:opacity-100">
                {externalLinks.map((site) => (
                  <Tooltip
                    key={site.key}
                    title={zh(`在 ${site.name} 中打开`, `Open in ${site.name}`)}
                    placement="top"
                    arrow
                    TransitionComponent={Fade}
                    TransitionProps={{ timeout: 0 }}
                  >
                    <a
                      href={site.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex h-7 w-7 items-center justify-center rounded-full bg-gray-100 transition hover:bg-gray-200"
                      aria-label={zh(`在 ${site.name} 中打开`, `Open in ${site.name}`)}
                      onClick={(event) => event.stopPropagation()}
                    >
                      <img src={site.icon} alt={site.name} className="h-4 w-4" loading="lazy" />
                    </a>
                  </Tooltip>
                ))}
              </div>
            </div>
          )}
          <Tooltip title={zh('用默认程序打开', 'Open with default app')}>
            <IconButton
              size="small"
              onClick={handleOpenFile}
              disabled={!canOpen}
              aria-label={zh('打开文件', 'Open file')}
              className="h-6 w-6"
            >
              <OpenInNewIcon fontSize="inherit" />
            </IconButton>
          </Tooltip>
          <Tooltip title={zh('打开所在位置', 'Reveal in folder')}>
            <IconButton
              size="small"
              onClick={handleRevealFile}
              disabled={!canOpen}
              aria-label={zh('打开所在位置', 'Reveal in folder')}
              className="h-6 w-6"
            >
              <FolderOpenIcon fontSize="inherit" />
            </IconButton>
          </Tooltip>
          {showEditTags && (
            <Tooltip title={zh('编辑标签', 'Edit tags')}>
              <IconButton
                size="small"
                onClick={handleEditTags}
                aria-label={zh('编辑标签', 'Edit tags')}
                className="h-6 w-6"
              >
                <EditIcon fontSize="inherit" />
              </IconButton>
            </Tooltip>
          )}
        </div>
      </div>
    </div>
  )
}
