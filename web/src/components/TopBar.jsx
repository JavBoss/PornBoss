import { useEffect, useRef } from 'react'
import { Button } from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import LocalOfferOutlinedIcon from '@mui/icons-material/LocalOfferOutlined'
import ShuffleOutlinedIcon from '@mui/icons-material/ShuffleOutlined'
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined'
import SwapHorizOutlinedIcon from '@mui/icons-material/SwapHorizOutlined'
import ArrowForwardRoundedIcon from '@mui/icons-material/ArrowForwardRounded'
import { zh } from '@/utils/i18n'

export default function TopBar({
  onHome,
  isJavMode,
  onToggleMode,
  videoSearchInput,
  onVideoSearchInputChange,
  onSubmitVideoSearch,
  videoSearchHref,
  randomHref,
  onRandomClick,
  onOpenTagModal,
  onOpenJavTagModal,
  onOpenVideoSettings,
  onOpenJavSettings,
  onOpenGlobalSettings,
  javSearchInput,
  onJavSearchInputChange,
  onSubmitJavSearch,
  javSearchHref,
  javRandomHref,
  javRandomMode,
  onJavRandomClick,
  isModifiedClick,
  javTab,
  onSwitchJavTab,
  filterSummary,
  showDirectorySetupHint,
}) {
  const headerRef = useRef(null)
  const headerClassName = ['sticky top-0 z-40 border-b bg-white/80 backdrop-blur']
    .filter(Boolean)
    .join(' ')

  const updateTopbarOffset = () => {
    const height = headerRef.current?.getBoundingClientRect().height || 0
    document.documentElement.style.setProperty('--topbar-height', `${Math.round(height)}px`)
  }

  useEffect(() => {
    updateTopbarOffset()
    window.addEventListener('resize', updateTopbarOffset)
    return () => window.removeEventListener('resize', updateTopbarOffset)
  }, [])

  useEffect(() => {
    updateTopbarOffset()
  }, [isJavMode, javTab, javRandomMode])

  const handleSettingsClick = () => {
    if (isJavMode) {
      onOpenJavSettings?.()
    } else {
      onOpenVideoSettings?.()
    }
  }

  const filterLabelPrefix = zh('筛选条件：', 'Filters:')

  return (
    <header ref={headerRef} className={headerClassName}>
      <div className="mx-auto flex max-w-screen-2xl flex-col gap-3 px-6 py-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex min-w-0 flex-1 flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={onHome}
              className="cursor-pointer select-none rounded text-left text-xl font-bold focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
            >
              Pornboss
            </button>

            <div className="flex min-w-0 flex-1 items-center gap-2">
              <div className="flex flex-wrap items-center gap-2">
                {isJavMode ? (
                  <div className="flex items-center gap-2">
                    <Button
                      variant={javTab === 'list' ? 'contained' : 'outlined'}
                      onClick={() => onSwitchJavTab('list')}
                    >
                      {zh('作品', 'JAV')}
                    </Button>
                    <Button
                      variant={javTab === 'idol' ? 'contained' : 'outlined'}
                      onClick={() => onSwitchJavTab('idol')}
                    >
                      {zh('女优', 'Idol')}
                    </Button>
                    <form
                      onSubmit={onSubmitJavSearch}
                      className="flex items-center overflow-hidden rounded-full border border-gray-200 bg-white shadow-sm"
                    >
                      <input
                        value={javSearchInput}
                        onChange={(e) => onJavSearchInputChange(e.target.value)}
                        placeholder={
                          javTab === 'idol'
                            ? zh('搜索女优名称', 'Search idol name')
                            : zh('搜索番号或标题', 'Search code or title')
                        }
                        className="h-10 flex-1 border-0 bg-white px-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        aria-label={zh('搜索JAV', 'Search JAV')}
                      />
                      <Button
                        component="a"
                        href={javSearchHref}
                        aria-label={zh('搜索JAV', 'Search JAV')}
                        variant="contained"
                        size="medium"
                        onClick={(e) => {
                          if (isModifiedClick(e)) return
                          onSubmitJavSearch(e)
                        }}
                        sx={{
                          borderTopLeftRadius: 0,
                          borderBottomLeftRadius: 0,
                          minHeight: '40px',
                          height: '40px',
                          px: 2.5,
                        }}
                      >
                        <SearchIcon fontSize="small" />
                      </Button>
                    </form>
                    <Button
                      component="a"
                      href={javRandomHref}
                      startIcon={<ShuffleOutlinedIcon fontSize="small" />}
                      variant="outlined"
                      onClick={(e) => {
                        if (isModifiedClick(e)) return
                        e.preventDefault()
                        onJavRandomClick?.()
                      }}
                    >
                      {zh('随机', 'Random')}
                    </Button>
                    <Button
                      startIcon={<LocalOfferOutlinedIcon fontSize="small" />}
                      variant="outlined"
                      onClick={onOpenJavTagModal}
                    >
                      {zh('标签', 'Tag')}
                    </Button>
                  </div>
                ) : (
                  <>
                    <form
                      onSubmit={onSubmitVideoSearch}
                      className="flex items-center overflow-hidden rounded-full border border-gray-200 bg-white shadow-sm"
                    >
                      <input
                        value={videoSearchInput}
                        onChange={(e) => onVideoSearchInputChange(e.target.value)}
                        placeholder={zh('搜索文件名', 'Search filename')}
                        className="h-10 flex-1 border-0 bg-white px-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        aria-label={zh('搜索视频', 'Search videos')}
                      />
                      <Button
                        component="a"
                        href={videoSearchHref}
                        aria-label={zh('搜索视频', 'Search videos')}
                        variant="contained"
                        size="medium"
                        onClick={(e) => {
                          if (isModifiedClick(e)) return
                          onSubmitVideoSearch(e)
                        }}
                        sx={{
                          borderTopLeftRadius: 0,
                          borderBottomLeftRadius: 0,
                          minHeight: '40px',
                          height: '40px',
                          px: 2.5,
                        }}
                      >
                        <SearchIcon fontSize="small" />
                      </Button>
                    </form>
                    <div className="flex items-center gap-2">
                      <Button
                        component="a"
                        href={randomHref}
                        startIcon={<ShuffleOutlinedIcon fontSize="small" />}
                        variant="outlined"
                        onClick={(e) => {
                          if (isModifiedClick(e)) return
                          e.preventDefault()
                          onRandomClick()
                        }}
                      >
                        {zh('随机', 'Random')}
                      </Button>
                    </div>
                    <Button
                      startIcon={<LocalOfferOutlinedIcon fontSize="small" />}
                      variant="outlined"
                      onClick={onOpenTagModal}
                    >
                      {zh('标签', 'Tag')}
                    </Button>
                  </>
                )}

                <Button
                  startIcon={<SettingsOutlinedIcon fontSize="small" />}
                  variant="outlined"
                  onClick={handleSettingsClick}
                  title={zh('全局设置', 'Global settings')}
                >
                  {zh('设置', 'Settings')}
                </Button>
              </div>

              {filterSummary ? (
                <span
                  className="min-w-0 flex-1 truncate whitespace-nowrap text-xs text-gray-500"
                  title={`${filterLabelPrefix}${filterSummary}`}
                >
                  {filterLabelPrefix}
                  <span className="font-semibold text-gray-700">{filterSummary}</span>
                </span>
              ) : null}
            </div>
          </div>

          <div className="flex flex-shrink-0 flex-wrap items-center justify-end gap-2">
            {showDirectorySetupHint ? (
              <div
                className="directory-setup-hint flex max-w-full items-center gap-2 rounded-full border border-amber-200 bg-amber-50 px-3 py-1.5 text-xs font-medium text-amber-900 shadow-sm"
                role="status"
              >
                <span className="min-w-0 truncate">
                  {zh(
                    '没有可用的视频目录，打开目录管理添加',
                    'No video directories. Open Directory Management to add one'
                  )}
                </span>
                <ArrowForwardRoundedIcon
                  className="directory-setup-hint__arrow shrink-0"
                  fontSize="small"
                  aria-hidden="true"
                />
              </div>
            ) : null}
            <Button
              startIcon={<SettingsOutlinedIcon fontSize="small" />}
              variant="outlined"
              onClick={onOpenGlobalSettings}
              title={zh('全局设置', 'Global settings')}
            >
              {zh('全局设置', 'Global Settings')}
            </Button>
            <Button
              variant="contained"
              color={isJavMode ? 'secondary' : 'primary'}
              startIcon={<SwapHorizOutlinedIcon fontSize="small" />}
              onClick={onToggleMode}
            >
              {isJavMode ? zh('切换到视频', 'To Video') : zh('切换到 JAV', 'To JAV')}
            </Button>
          </div>
        </div>
      </div>
    </header>
  )
}
