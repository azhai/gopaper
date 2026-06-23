import m from 'mithril';
import { apiService } from '../services/apiService';
import { showAlert } from '../components/overlay';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
import { topbarState } from '../components/navbar';
export const EditorPage = {
    oninit(vnode) {
        const isEdit = !!vnode.attrs.slug;
        const isPage = vnode.attrs.mode === 'page';
        Object.assign(vnode.state, {
            isEdit,
            isPage,
            slug: vnode.attrs.slug || '',
            form: {
                dirPath: '', title: '', slug: '', author: '',
                date: new Date().toISOString().split('T')[0],
                tags: '', weight: 0, position: '', comments: true, content: '',
            },
            previewHtml: '',
            showPreview: false,
            saving: false,
            loading: isEdit,
            dirs: [],
            regions: [],
        });
        const noun = isPage ? '页面' : '文章';
        topbarState.title = isEdit ? `编辑${noun}` : `新建${noun}`;
        topbarState.crumb = isEdit ? (vnode.attrs.slug || null) : null;
        topbarState.actions = null;
    },
    async oncreate(vnode) {
        const s = vnode.state;
        apiService.getDirs().then(res => {
            const all = (res.data || []);
            s.dirs = s.isPage ? all.filter(d => d.dirType === 'page') : all.filter(d => d.dirType !== 'page');
            m.redraw();
        }).catch(() => { });
        // Load layout regions for position selector
        apiService.getLayouts().then(res => {
            if (res.code === 0 && res.data) {
                const templates = res.data.templates || [];
                // Collect all unique regions across templates
                const seen = new Set();
                s.regions = [];
                for (const t of templates) {
                    for (const r of (t.regions || [])) {
                        if (!seen.has(r.name)) {
                            seen.add(r.name);
                            s.regions.push(r);
                        }
                    }
                }
                m.redraw();
            }
        }).catch(() => { });
        if (s.isEdit && s.slug) {
            try {
                const result = await apiService.getArticle(s.slug);
                if (result.code === 0 && result.data) {
                    const d = result.data;
                    s.form = {
                        dirPath: d.dirPath || '',
                        title: d.title || '',
                        slug: d.slug || '',
                        author: d.author || '',
                        date: d.date || '',
                        tags: (d.tags || []).join(', '),
                        weight: d.weight || 0,
                        position: d.position || '',
                        comments: d.comments !== false,
                        content: d.content || '',
                    };
                }
            }
            finally {
                s.loading = false;
                m.redraw();
            }
        }
        else {
            s.loading = false;
        }
    },
    view(vnode) {
        const s = vnode.state;
        const noun = s.isPage ? '页面' : '文章';
        if (s.loading)
            return m('.loading-state', [m('.spinner'), m('p', `加载${noun}内容...`)]);
        const listHref = s.isPage ? '/pages' : '/articles';
        return m('.editor-page', [
            m('.editor-layout', { class: s.showPreview ? 'with-preview' : '' }, [
                m('div', [
                    // Meta card
                    m('.card.editor-meta-card', [
                        m('.section-title', [Icon('info'), s.isPage ? '页面信息' : '文章信息']),
                        m('.form-row.cols-2', [
                            field(s.isPage ? '位置' : '目录路径', m('select.form-control', {
                                value: s.form.dirPath,
                                onchange: (e) => { s.form.dirPath = e.target.value; },
                            }, [
                                m('option', { value: '' }, '/ (根目录)'),
                                ...s.dirs.map(d => m('option', { value: d.dirPath }, `${d.title || d.dirPath} (${d.dirPath})`)),
                            ])),
                            field('标题', true, m('input.form-control', {
                                type: 'text', value: s.form.title, placeholder: s.isPage ? '请输入页面标题' : '请输入文章标题',
                                oninput: (e) => { s.form.title = e.target.value; },
                            })),
                        ]),
                        s.isPage
                            ? m('.form-row.cols-2', [
                                field('Slug', m('input.form-control', {
                                    type: 'text', value: s.form.slug, placeholder: '留空自动生成',
                                    oninput: (e) => { s.form.slug = e.target.value; },
                                })),
                                field('排序 (权重)', m('input.form-control', {
                                    type: 'number', value: String(s.form.weight),
                                    oninput: (e) => { s.form.weight = parseInt(e.target.value) || 0; },
                                })),
                            ])
                            : null,
                        // Position selector (layout region)
                        m('.form-row.cols-2', [
                            field('布局位置', m('select.form-control', {
                                value: s.form.position,
                                onchange: (e) => { s.form.position = e.target.value; },
                            }, [
                                m('option', { value: '' }, '默认 (main)'),
                                ...s.regions.map(r => m('option', { value: r.name }, `${r.title} (${r.name})`)),
                            ])),
                            field('排序 (权重)', m('input.form-control', {
                                type: 'number', value: String(s.form.weight),
                                oninput: (e) => { s.form.weight = parseInt(e.target.value) || 0; },
                                disabled: s.isPage,
                            })),
                        ]),
                        !s.isPage
                            ? m('.form-row.cols-3', [
                                field('Slug', m('input.form-control', {
                                    type: 'text', value: s.form.slug, placeholder: '留空自动生成',
                                    oninput: (e) => { s.form.slug = e.target.value; },
                                })),
                                field('作者', m('input.form-control', {
                                    type: 'text', value: s.form.author, placeholder: '作者',
                                    oninput: (e) => { s.form.author = e.target.value; },
                                })),
                                field('日期', m('input.form-control', {
                                    type: 'date', value: s.form.date,
                                    oninput: (e) => { s.form.date = e.target.value; },
                                })),
                            ])
                            : null,
                        !s.isPage
                            ? m('.form-row.cols-2-1', [
                                field('标签', m('input.form-control', {
                                    type: 'text', value: s.form.tags, placeholder: '用逗号分隔多个标签',
                                    oninput: (e) => { s.form.tags = e.target.value; },
                                })),
                                field('权重', m('input.form-control', {
                                    type: 'number', value: String(s.form.weight),
                                    oninput: (e) => { s.form.weight = parseInt(e.target.value) || 0; },
                                })),
                            ])
                            : null,
                        !s.isPage
                            ? m('label.checkbox-row', [
                                m('input', {
                                    type: 'checkbox', checked: s.form.comments,
                                    onchange: (e) => { s.form.comments = e.target.checked; },
                                }),
                                m('span', { style: { fontSize: '14px', color: 'var(--c-text)' } }, '允许评论'),
                            ])
                            : null,
                    ]),
                    // Content card
                    m('.card.editor-content-card', { style: { marginTop: '20px' } }, [
                        m('.editor-toolbar', [
                            m('.editor-toolbar-left', [
                                m('.section-title', { style: { marginBottom: '0' } }, [Icon('edit'), '正文 (Markdown)']),
                            ]),
                            m('div', { style: { display: 'flex', gap: '8px' } }, [
                                m('button.btn.btn-ghost.btn-sm', {
                                    onclick: async () => {
                                        if (!s.showPreview && s.form.content) {
                                            const result = await apiService.previewArticle(s.form.content);
                                            if (result.code === 0 && result.data) {
                                                s.previewHtml = result.data.html;
                                            }
                                        }
                                        s.showPreview = !s.showPreview;
                                    },
                                }, [Icon('preview'), s.showPreview ? '隐藏预览' : '显示预览']),
                            ]),
                        ]),
                        m('textarea.editor-textarea', {
                            value: s.form.content,
                            placeholder: '# 在此输入 Markdown 内容...\n\n支持标题、列表、代码块、图片等。',
                            oninput: (e) => { s.form.content = e.target.value; },
                        }),
                    ]),
                    m('div', { style: { display: 'flex', gap: '12px', marginTop: '20px' } }, [
                        m('button.btn.btn-primary', {
                            disabled: s.saving,
                            onclick: async () => {
                                if (!s.form.title || !s.form.content) {
                                    showAlert('校验失败', '标题和正文不能为空');
                                    return;
                                }
                                s.saving = true;
                                m.redraw();
                                try {
                                    const data = {
                                        dirPath: s.form.dirPath || '/',
                                        title: s.form.title,
                                        slug: s.form.slug || undefined,
                                        author: s.isPage ? undefined : (s.form.author || undefined),
                                        date: s.isPage ? undefined : (s.form.date || undefined),
                                        tags: s.isPage ? undefined : (s.form.tags ? s.form.tags.split(',').map(t => t.trim()).filter(Boolean) : undefined),
                                        weight: s.form.weight || undefined,
                                        position: s.form.position || undefined,
                                        comments: s.isPage ? undefined : s.form.comments,
                                        content: s.form.content,
                                    };
                                    const result = s.isEdit
                                        ? await apiService.updateArticle(s.slug, data)
                                        : await apiService.createArticle(data);
                                    if (result.code === 0) {
                                        toastService.success(s.isEdit ? '更新成功' : '创建成功');
                                        if (!s.isEdit)
                                            m.route.set(listHref);
                                    }
                                    else {
                                        showAlert('保存失败', result.message || '未知错误');
                                    }
                                }
                                catch {
                                    showAlert('保存失败', '网络错误');
                                }
                                finally {
                                    s.saving = false;
                                    m.redraw();
                                }
                            },
                        }, [Icon('save'), s.saving ? '保存中...' : `保存${noun}`]),
                        m(m.route.Link, { href: listHref, class: 'btn btn-ghost' }, '取消'),
                    ]),
                ]),
                s.showPreview ? m('.preview-panel', [
                    m('.section-title', [Icon('eye'), '预览']),
                    m('.preview-content', { style: { lineHeight: 1.9 }, innerHTML: s.previewHtml || '<p style="color:#c4a35a">点击「显示预览」按钮渲染内容</p>' }),
                ]) : null,
            ]),
        ]);
    },
};
function field(label, requiredOrInput, input) {
    const required = typeof requiredOrInput === 'boolean' ? requiredOrInput : false;
    const node = typeof requiredOrInput === 'boolean' ? input : requiredOrInput;
    return m('.form-group', [
        m('label.form-label', [label, required ? m('span.req', '*') : null]),
        node,
    ]);
}
