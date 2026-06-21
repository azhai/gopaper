import m from 'mithril';
import { apiService } from '../services/apiService';
import { showDangerConfirm, showAlert } from '../components/overlay';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
export const ImagesPage = {
    oninit(vnode) {
        Object.assign(vnode.state, { images: [], total: 0, page: 1, pageSize: 20, loading: true, dragging: false, lightbox: null });
    },
    oncreate(vnode) {
        loadImages(vnode);
    },
    view(vnode) {
        const s = vnode.state;
        return m('.images-page', [
            m('.page-header', [
                m('div', [
                    m('h2', '图片管理'),
                    m('.desc', `共 ${s.total} 张图片`),
                ]),
            ]),
            m('.upload-zone', {
                class: s.dragging ? 'dragging' : '',
                ondragover: (e) => { e.preventDefault(); s.dragging = true; },
                ondragleave: () => { s.dragging = false; },
                ondrop: async (e) => {
                    e.preventDefault();
                    s.dragging = false;
                    const files = e.dataTransfer?.files;
                    if (files && files.length > 0) {
                        await uploadFile(vnode, files[0]);
                    }
                },
                onclick: () => {
                    const input = document.createElement('input');
                    input.type = 'file';
                    input.accept = 'image/*';
                    input.onchange = async () => {
                        if (input.files && input.files.length > 0) {
                            await uploadFile(vnode, input.files[0]);
                        }
                    };
                    input.click();
                },
            }, [
                Icon('image-plus'),
                m('h4', '拖拽图片到此处或点击上传'),
                m('p', '支持 jpg / png / gif / webp / svg，最大 5MB'),
            ]),
            s.loading
                ? m('.loading-state', [m('.spinner'), m('p', '加载中...')])
                : s.images.length === 0
                    ? m('.empty-state', [Icon('image'), m('h3', '暂无图片'), m('p', '上传第一张图片吧')])
                    : m('.image-grid', s.images.map(img => m('.image-card', { key: img.fileName }, [
                        m('img.image-thumb', {
                            src: img.uploadPath,
                            loading: 'lazy',
                            onclick: () => { s.lightbox = img; },
                        }),
                        m('.image-info', [
                            m('.image-name', { title: img.originalName }, img.originalName),
                            m('.image-meta', [
                                m('span', formatSize(img.fileSize)),
                                m('span', img.fileType.replace('image/', '')),
                            ]),
                            m('.image-actions', [
                                m('button.btn.btn-ghost.btn-sm', {
                                    style: { flex: '1' },
                                    title: '复制链接',
                                    onclick: async () => {
                                        try {
                                            await navigator.clipboard.writeText(window.location.origin + img.uploadPath);
                                            toastService.success('链接已复制');
                                        }
                                        catch {
                                            toastService.error('复制失败');
                                        }
                                    },
                                }, Icon('copy', { style: { width: '14px', height: '14px' } })),
                                m('button.btn.btn-danger.btn-sm', {
                                    title: '删除',
                                    onclick: () => {
                                        showDangerConfirm('删除图片', `确定要删除「${img.originalName}」吗？`, async () => {
                                            const result = await apiService.deleteImage(img.fileName);
                                            if (result.code === 0) {
                                                toastService.success('删除成功');
                                                loadImages(vnode);
                                            }
                                            else if (result.code === 40902) {
                                                showAlert('无法删除', result.message || '该图片仍被文章引用');
                                            }
                                            else {
                                                showAlert('删除失败', result.message || '未知错误');
                                            }
                                        });
                                    },
                                }, Icon('trash', { style: { width: '14px', height: '14px' } })),
                            ]),
                        ]),
                    ]))),
            s.lightbox ? m('.lightbox', {
                onclick: () => { s.lightbox = null; },
            }, [
                m('img', { src: s.lightbox.uploadPath, onclick: (e) => e.stopPropagation() }),
                m('div', {
                    style: { position: 'absolute', top: '24px', right: '24px', color: '#fff', cursor: 'pointer', background: 'rgba(0,0,0,.4)', borderRadius: '50%', width: '40px', height: '40px', display: 'flex', alignItems: 'center', justifyContent: 'center' },
                    onclick: () => { s.lightbox = null; },
                }, Icon('x', { style: { width: '22px', height: '22px' } })),
            ]) : null,
        ]);
    },
};
async function loadImages(vnode) {
    vnode.state.loading = true;
    try {
        const result = await apiService.getImages(vnode.state.page, vnode.state.pageSize);
        if (result.code === 0) {
            vnode.state.images = (result.data || []);
            vnode.state.total = result.total || 0;
        }
    }
    finally {
        vnode.state.loading = false;
        m.redraw();
    }
}
async function uploadFile(vnode, file) {
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'];
    if (!validTypes.includes(file.type)) {
        showAlert('上传失败', '不支持的文件类型，仅允许 jpg/png/gif/webp/svg');
        return;
    }
    if (file.size > 5 * 1024 * 1024) {
        showAlert('上传失败', '文件大小超过 5MB 限制');
        return;
    }
    try {
        const result = await apiService.uploadImage(file);
        if (result.code === 0) {
            toastService.success('上传成功');
            loadImages(vnode);
        }
        else {
            showAlert('上传失败', result.message || '未知错误');
        }
    }
    catch {
        showAlert('上传失败', '网络错误');
    }
}
function formatSize(bytes) {
    if (bytes < 1024)
        return bytes + ' B';
    if (bytes < 1024 * 1024)
        return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
}
