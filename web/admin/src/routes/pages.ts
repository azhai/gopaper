import m from 'mithril';
import { apiService, DirInfo } from '../services/apiService';
import { showDangerConfirm, showAlert } from '../components/overlay';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
import { topbarState } from '../components/navbar';

interface PageSummary {
  slug: string;
  title: string;
  weight?: number;
  dirPath: string;
}

interface PagesState {
  pages: PageSummary[];
  total: number;
  loading: boolean;
  dirs: DirInfo[];
  keyword: string;
}

export const PagesPage: m.Component<{}, PagesState> = {
  oninit(vnode) {
    Object.assign(vnode.state, {
      pages: [], total: 0, loading: true, dirs: [], keyword: '',
    });
    topbarState.title = '页面管理';
    topbarState.crumb = null;
    topbarState.actions = null;
  },

  oncreate(vnode) {
    apiService.getDirs().then(res => {
      vnode.state.dirs = ((res.data || []) as DirInfo[]).filter(d => d.dirType === 'page');
      m.redraw();
    }).catch(() => {});
    loadPages(vnode);
  },

  view(vnode) {
    const s = vnode.state;

    return m('.pages-page', [
      m('.page-header', [
        m('div', [
          m('h2', '页面管理'),
          m('.desc', `共 ${s.total} 个页面 · 页面无作者/日期/标签，按位置与排序展示`),
        ]),
        m('.page-header-actions', [
          m(m.route.Link, { href: '/pages/new', class: 'btn btn-primary' }, [Icon('plus'), '新建页面']),
        ]),
      ]),

      m('.toolbar', [
        m('.form-control', {
          style: { position: 'relative', display: 'flex', alignItems: 'center', minWidth: '240px', padding: 0, border: '1px solid var(--c-border-soft)', borderRadius: 'var(--radius)', background: 'var(--c-card)' },
        }, [
          m('span', { style: { padding: '0 10px', color: 'var(--c-text-muted)', display: 'flex' } }, Icon('search', { style: { width: '16px', height: '16px' } })),
          m('input', {
            placeholder: '搜索页面标题...',
            value: s.keyword,
            style: { border: 'none', outline: 'none', background: 'transparent', padding: '9px 12px 9px 0', flex: '1', fontSize: '14px', color: 'var(--c-text)' },
            oninput: (e: Event) => {
              s.keyword = (e.target as HTMLInputElement).value;
              m.redraw();
            },
          }),
        ]),
      ]),

      s.loading
        ? m('.loading-state', [m('.spinner'), m('p', '加载中...')])
        : filteredPages(s).length === 0
          ? m('.empty-state', [Icon('file'), m('h3', '暂无页面'), m('p', '点击右上角「新建页面」创建站点页面')])
          : m('.table-wrap', m('table.data-table', [
              m('thead', m('tr', [
                m('th', '页面标题'),
                m('th', '位置'),
                m('th', { style: { width: '100px' } }, '排序'),
                m('th', { style: { width: '140px' } }, '操作'),
              ])),
              m('tbody', filteredPages(s).map(page =>
                m('tr', { key: page.slug }, [
                  m('td', m('.cell-title', m(m.route.Link, { href: `/pages/${page.slug}/edit` }, page.title))),
                  m('td', m('span.tag-pill.dir', dirLabel(s.dirs, page.dirPath))),
                  m('td', m('.cell-muted', page.weight || 0)),
                  m('td', m('.cell-actions', [
                    m('a.btn.btn-ghost.btn-sm', {
                      href: '/' + (page.dirPath || '') + (page.dirPath ? '/' : '') + page.slug,
                      target: '_blank',
                      title: '前台查看',
                    }, Icon('external', { style: { width: '15px', height: '15px' } })),
                    m(m.route.Link, {
                      href: `/pages/${page.slug}/edit`,
                      class: 'btn btn-ghost btn-sm',
                      title: '编辑',
                    }, Icon('edit', { style: { width: '15px', height: '15px' } })),
                    m('button.btn.btn-danger.btn-sm', {
                      title: '删除',
                      onclick: () => {
                        showDangerConfirm('删除页面', `确定要删除页面「${page.title}」吗？此操作不可恢复。`, async () => {
                          const result = await apiService.deleteArticle(page.slug);
                          if (result.code === 0) {
                            toastService.success('删除成功');
                            loadPages(vnode);
                          } else {
                            showAlert('删除失败', result.message || '未知错误');
                          }
                        });
                      },
                    }, Icon('trash', { style: { width: '15px', height: '15px' } })),
                  ])),
                ]),
              )),
            ])),
    ]);
  },
};

function dirLabel(dirs: DirInfo[], dirPath: string): string {
  if (!dirPath) return '/ (首页)';
  const d = dirs.find(x => x.dirPath === dirPath);
  return d ? `${d.title} (${d.dirPath})` : dirPath;
}

function filteredPages(s: PagesState): PageSummary[] {
  if (!s.keyword.trim()) return s.pages;
  const kw = s.keyword.trim().toLowerCase();
  return s.pages.filter(p =>
    p.title.toLowerCase().includes(kw) ||
    p.dirPath.toLowerCase().includes(kw),
  );
}

async function loadPages(vnode: m.Vnode<{}, PagesState>) {
  const s = vnode.state;
  s.loading = true;
  try {
    const result = await apiService.getArticles(1, 100, undefined, 'page');
    if (result.code === 0) {
      s.pages = (result.data || []) as PageSummary[];
      s.total = result.total || 0;
    }
  } finally {
    s.loading = false;
    m.redraw();
  }
}
