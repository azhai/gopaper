import m from 'mithril';
import { apiService } from '../services/apiService';
import { showDangerConfirm, showAlert } from '../components/overlay';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
export const ArticlesPage = {
    oninit(vnode) {
        Object.assign(vnode.state, {
            articles: [], total: 0, page: 1, pageSize: 10, loading: true,
            dirs: [], dirFilter: '', keyword: '',
        });
    },
    oncreate(vnode) {
        apiService.getDirs().then(res => {
            vnode.state.dirs = (res.data || []).filter(d => d.dirType !== 'page');
            m.redraw();
        }).catch(() => { });
        loadArticles(vnode);
    },
    view(vnode) {
        const s = vnode.state;
        const totalPages = Math.ceil(s.total / s.pageSize);
        return m('.articles-page', [
            m('.page-header', [
                m('div', [
                    m('h2', '文章管理'),
                    m('.desc', `共 ${s.total} 篇文章`),
                ]),
                m('.page-header-actions', [
                    m(m.route.Link, { href: '/articles/new', class: 'btn btn-primary' }, [Icon('plus'), '新建文章']),
                ]),
            ]),
            m('.toolbar', [
                m('.form-control', {
                    style: { position: 'relative', display: 'flex', alignItems: 'center', minWidth: '240px', padding: 0, border: '1px solid var(--c-border-soft)', borderRadius: 'var(--radius)', background: 'var(--c-card)' },
                }, [
                    m('span', { style: { padding: '0 10px', color: 'var(--c-text-muted)', display: 'flex' } }, Icon('search', { style: { width: '16px', height: '16px' } })),
                    m('input', {
                        placeholder: '搜索标题...',
                        value: s.keyword,
                        style: { border: 'none', outline: 'none', background: 'transparent', padding: '9px 12px 9px 0', flex: '1', fontSize: '14px', color: 'var(--c-text)' },
                        oninput: (e) => {
                            s.keyword = e.target.value;
                            s.page = 1;
                            loadArticles(vnode);
                        },
                    }),
                ]),
                m('select.form-control', {
                    style: { minWidth: '160px' },
                    value: s.dirFilter,
                    onchange: (e) => {
                        s.dirFilter = e.target.value;
                        s.page = 1;
                        loadArticles(vnode);
                    },
                }, [
                    m('option', { value: '' }, '全部目录'),
                    ...s.dirs.map(d => m('option', { value: d.dirPath }, d.title || d.dirPath)),
                ]),
            ]),
            s.loading
                ? m('.loading-state', [m('.spinner'), m('p', '加载中...')])
                : s.articles.length === 0
                    ? m('.empty-state', [Icon('article'), m('h3', '暂无文章'), m('p', '点击右上角「新建文章」开始创作')])
                    : m('.table-wrap', m('table.data-table', [
                        m('thead', m('tr', [
                            m('th', '标题'),
                            m('th', '作者'),
                            m('th', '日期'),
                            m('th', '目录'),
                            m('th', '标签'),
                            m('th', { style: { width: '120px' } }, '操作'),
                        ])),
                        m('tbody', filteredArticles(s).map(article => m('tr', { key: article.slug }, [
                            m('td', m('.cell-title', m(m.route.Link, { href: `/articles/${article.slug}/edit` }, article.title))),
                            m('td', m('.cell-muted', article.author || '-')),
                            m('td', m('.cell-muted', article.date || '-')),
                            m('td', m('span.tag-pill.dir', article.dirPath || '/')),
                            m('td', (article.tags || []).map(t => m('span.tag-pill', t))),
                            m('td', m('.cell-actions', [
                                m(m.route.Link, {
                                    href: `/articles/${article.slug}/edit`,
                                    class: 'btn btn-ghost btn-sm',
                                    title: '编辑',
                                }, Icon('edit', { style: { width: '15px', height: '15px' } })),
                                m('button.btn.btn-danger.btn-sm', {
                                    title: '删除',
                                    onclick: () => {
                                        showDangerConfirm('删除文章', `确定要删除「${article.title}」吗？此操作不可恢复。`, async () => {
                                            const result = await apiService.deleteArticle(article.slug);
                                            if (result.code === 0) {
                                                toastService.success('删除成功');
                                                loadArticles(vnode);
                                            }
                                            else {
                                                showAlert('删除失败', result.message || '未知错误');
                                            }
                                        });
                                    },
                                }, Icon('trash', { style: { width: '15px', height: '15px' } })),
                            ])),
                        ]))),
                    ])),
            totalPages > 1 ? m('.pagination', [
                m('button.page-btn', {
                    disabled: s.page <= 1,
                    onclick: () => { s.page--; loadArticles(vnode); },
                }, Icon('chevron-left', { style: { width: '16px', height: '16px' } })),
                ...pageRange(s.page, totalPages).map(p => p === '...' ? m('span', { style: { padding: '0 6px', color: 'var(--c-text-muted)' } }, '...') :
                    m('button.page-btn', {
                        class: p === s.page ? 'active' : '',
                        onclick: () => { s.page = p; loadArticles(vnode); },
                    }, String(p))),
                m('button.page-btn', {
                    disabled: s.page >= totalPages,
                    onclick: () => { s.page++; loadArticles(vnode); },
                }, Icon('chevron-right', { style: { width: '16px', height: '16px' } })),
            ]) : null,
        ]);
    },
};
function filteredArticles(s) {
    if (!s.keyword.trim())
        return s.articles;
    const kw = s.keyword.trim().toLowerCase();
    return s.articles.filter(a => a.title.toLowerCase().includes(kw) ||
        (a.author || '').toLowerCase().includes(kw));
}
function pageRange(current, total) {
    const range = [];
    const delta = 2;
    const left = Math.max(1, current - delta);
    const right = Math.min(total, current + delta);
    if (left > 1) {
        range.push(1);
        if (left > 2)
            range.push('...');
    }
    for (let i = left; i <= right; i++)
        range.push(i);
    if (right < total) {
        if (right < total - 1)
            range.push('...');
        range.push(total);
    }
    return range;
}
async function loadArticles(vnode) {
    const s = vnode.state;
    s.loading = true;
    try {
        const result = await apiService.getArticles(s.page, s.pageSize, s.dirFilter || undefined, 'article');
        if (result.code === 0) {
            s.articles = (result.data || []);
            s.total = result.total || 0;
        }
    }
    finally {
        s.loading = false;
        m.redraw();
    }
}
