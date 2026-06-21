import m from 'mithril';
import { apiService, DirInfo } from '../services/apiService';
import { toastService } from '../components/toast';
import { Icon } from '../icons';

interface RecentItem { slug: string; title: string; author?: string; date?: string; dirPath: string; }

interface DashboardState {
  articleCount: number;
  pageCount: number;
  imageCount: number;
  dirCount: number;
  dirs: DirInfo[];
  recentArticles: RecentItem[];
  recentPages: RecentItem[];
  loading: boolean;
  refreshing: boolean;
  builtAt: string;
}

export const DashboardPage: m.Component<{}, DashboardState> = {
  oninit(vnode) {
    Object.assign(vnode.state, {
      articleCount: 0, pageCount: 0, imageCount: 0, dirCount: 0,
      dirs: [], recentArticles: [], recentPages: [],
      loading: true, refreshing: false, builtAt: '',
    });
  },

  async oncreate(vnode) {
    const s = vnode.state;
    try {
      const [articlesRes, pagesRes, imagesRes, dirsRes] = await Promise.all([
        apiService.getArticles(1, 6, undefined, 'article'),
        apiService.getArticles(1, 6, undefined, 'page'),
        apiService.getImages(1, 1),
        apiService.getDirs(),
      ]);
      s.articleCount = articlesRes.total || 0;
      s.pageCount = pagesRes.total || 0;
      s.imageCount = imagesRes.total || 0;
      s.recentArticles = (articlesRes.data || []) as RecentItem[];
      s.recentPages = (pagesRes.data || []) as RecentItem[];
      s.dirs = (dirsRes.data || []) as DirInfo[];
      s.dirCount = s.dirs.length;
      m.redraw();
    } catch {
      // ignore
    } finally {
      s.loading = false;
      m.redraw();
    }
  },

  view(vnode) {
    const s = vnode.state;

    if (s.loading) {
      return m('.loading-state', [m('.spinner'), m('p', '正在加载仪表盘数据...')]);
    }

    return m('.dashboard', [
      m('.stat-grid', [
        statCard('gold', 'article', s.articleCount, '文章总数', '篇新闻/文档内容'),
        statCard('violet', 'file', s.pageCount, '页面总数', '个站点页面'),
        statCard('coral', 'image', s.imageCount, '图片总数', '已上传媒体文件'),
        statCard('green', 'folder', s.dirCount, '目录数量', '内容分类目录'),
      ]),

      m('.grid-2', { style: { marginTop: '24px', alignItems: 'start' } }, [
        // Recent articles
        m('.card', [
          m('.card-pad', { style: { paddingBottom: '0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' } }, [
            m('.section-title', { style: { marginBottom: '0' } }, [Icon('article'), '最近文章']),
            m(m.route.Link, { href: '/articles', class: 'btn-link' }, '全部'),
          ]),
          m('.dash-list', s.recentArticles.length > 0
            ? s.recentArticles.map(a => m('.dash-list-item', [
                m('.icon-circle', { style: iconBoxStyle('#fff8e1', 'var(--c-primary-dark)') }, Icon('article')),
                m('.title', m(m.route.Link, { href: `/articles/${a.slug}/edit` }, a.title)),
                m('.meta', a.date || '-'),
              ]))
            : [m('.empty-state', { style: { padding: '32px' } }, '暂无文章')]),
        ]),

        // Recent pages
        m('.card', [
          m('.card-pad', { style: { paddingBottom: '0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' } }, [
            m('.section-title', { style: { marginBottom: '0' } }, [Icon('file'), '站点页面']),
            m(m.route.Link, { href: '/pages', class: 'btn-link' }, '全部'),
          ]),
          m('.dash-list', s.recentPages.length > 0
            ? s.recentPages.map(p => m('.dash-list-item', [
                m('.icon-circle', { style: iconBoxStyle('#f0e7ff', '#7c3aed') }, Icon('file')),
                m('.title', m(m.route.Link, { href: `/pages/${p.slug}/edit` }, p.title)),
                m('.meta', p.dirPath || '/'),
              ]))
            : [m('.empty-state', { style: { padding: '32px' } }, '暂无页面')]),
        ]),
      ]),

      // Dirs overview
      m('.card', { style: { marginTop: '24px' } }, [
        m('.card-pad', { style: { paddingBottom: '0' } }, [
          m('.section-title', [Icon('folder'), '目录概览']),
        ]),
        m('.card-pad', { style: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: '10px' } },
          s.dirs.length > 0
            ? s.dirs.map(d => m(m.route.Link, {
                href: d.dirType === 'page' ? '/pages' : '/articles',
                style: { textDecoration: 'none' },
              }, m('.dir-chip', [
                m('.icon', { style: { background: d.dirType === 'page' ? '#f0e7ff' : '#fff8e1', color: d.dirType === 'page' ? '#7c3aed' : 'var(--c-primary-dark)' } }, Icon(d.dirType === 'page' ? 'file' : 'folder')),
                m('.info', [
                  m('.name', d.title || d.dirPath),
                  m('.count', `${d.articleCount} 个${d.dirType === 'page' ? '页面' : '文章'} · ${typeLabel(d.dirType)}`),
                ]),
                Icon('chevron-right', { style: { width: '16px', height: '16px' } }),
              ])))
            : [m('.empty-state', { style: { padding: '32px' } }, '暂无目录')],
        ),
      ]),

      // Quick actions
      m('.card', { style: { marginTop: '24px' } }, m('.card-pad', [
        m('.section-title', [Icon('zap'), '快捷操作']),
        m('div', { style: { display: 'flex', gap: '12px', flexWrap: 'wrap' } }, [
          m(m.route.Link, { href: '/articles/new', class: 'btn btn-primary' }, [Icon('plus'), '新建文章']),
          m(m.route.Link, { href: '/pages/new', class: 'btn btn-ghost' }, [Icon('plus'), '新建页面']),
          m(m.route.Link, { href: '/images', class: 'btn btn-ghost' }, [Icon('upload'), '上传图片']),
          m('button.btn.btn-ghost', {
            disabled: s.refreshing,
            onclick: async () => {
              s.refreshing = true;
              m.redraw();
              try {
                const result = await apiService.refreshCache();
                if (result.code === 0) {
                  s.builtAt = new Date().toLocaleTimeString('zh-CN');
                  toastService.success(`缓存刷新完成，共 ${result.data?.articleCount} 篇内容，耗时 ${result.data?.duration}`);
                } else {
                  toastService.error('缓存刷新失败');
                }
              } catch {
                toastService.error('网络错误');
              } finally {
                s.refreshing = false;
                m.redraw();
              }
            },
          }, [Icon('refresh'), s.refreshing ? '刷新中...' : '刷新缓存']),
          m(m.route.Link, { href: '/settings', class: 'btn btn-ghost' }, [Icon('settings'), '站点设置']),
        ]),
      ])),
    ]);
  },
};

function iconBoxStyle(bg: string, color: string): Record<string, string> {
  return { width: '32px', height: '32px', borderRadius: '8px', background: bg, color, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: '0' };
}

function statCard(color: string, icon: string, value: m.Children, label: string, sub: string): m.Vnode {
  return m(`.stat-card.${color}`, [
    m('.stat-card-top', [
      m(`.stat-icon.${color}`, Icon(icon as any)),
    ]),
    m('.stat-value', value),
    m('.stat-label', label),
    m('.stat-sub', sub),
  ]);
}

function typeLabel(t: string): string {
  switch (t) {
    case 'page': return '单页';
    case 'news': return '新闻';
    case 'docs': return '文档';
    default: return t || '未分类';
  }
}
