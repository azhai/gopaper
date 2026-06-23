import m from 'mithril';
import { authService } from '../services/apiService';
import { Icon } from '../icons';
import { toastService } from './toast';
export const topbarState = {
    title: null, crumb: null, actions: null,
};
const navItems = [
    { href: '/dashboard', label: '仪表盘', icon: 'dashboard' },
    { href: '/articles', label: '文章管理', icon: 'article' },
    { href: '/pages', label: '页面管理', icon: 'file' },
    { href: '/images', label: '图片管理', icon: 'image' },
    { href: '/layout', label: '布局管理', icon: 'layout' },
    { href: '/settings', label: '站点设置', icon: 'settings' },
];
let sidebarOpen = false;
function isActive(href) {
    const current = m.route.get();
    if (href === '/articles' || href === '/pages' || href === '/layout') {
        return current.startsWith(href);
    }
    return current === href;
}
export const Sidebar = {
    view() {
        if (!authService.isLoggedIn())
            return null;
        return [
            m('aside.sidebar', { class: sidebarOpen ? 'open' : '' }, [
                m('.sidebar-brand', [
                    m('.sidebar-brand-icon', m('img', { src: '/static/logo-mist.svg', alt: 'GoPaper', style: { width: '100%', height: '100%' } })),
                    m('div', [
                        m('.sidebar-brand-name', 'GoPaper'),
                        m('.sidebar-brand-sub', '内容管理后台'),
                    ]),
                ]),
                m('nav.sidebar-nav', [
                    m('.sidebar-section-label', '主导航'),
                    ...navItems.map(item => m(m.route.Link, {
                        href: item.href,
                        class: 'sidebar-link' + (isActive(item.href) ? ' active' : ''),
                        onclick: () => { sidebarOpen = false; },
                    }, [
                        Icon(item.icon),
                        item.label,
                    ])),
                ]),
                m('.sidebar-footer', [
                    m('.sidebar-user', [
                        m('.sidebar-user-avatar', 'A'),
                        m('.sidebar-user-info', [
                            m('.sidebar-user-name', 'Admin'),
                            m('.sidebar-user-role', '管理员'),
                        ]),
                        m('button.btn.btn-icon.btn-ghost', {
                            style: { color: 'var(--c-sidebar-text)', border: 'none', background: 'rgba(255,217,102,.1)' },
                            title: '刷新缓存',
                            onclick: async () => {
                                try {
                                    const res = await fetch('/api/admin/cache/refresh', {
                                        method: 'POST',
                                        headers: { 'Authorization': `Bearer ${authService.getToken()}` },
                                    });
                                    const data = await res.json();
                                    if (data.code === 0) {
                                        toastService.success(`缓存已刷新，共 ${data.data?.articleCount ?? 0} 篇文章`);
                                    }
                                    else {
                                        toastService.error('缓存刷新失败');
                                    }
                                }
                                catch {
                                    toastService.error('网络错误');
                                }
                            },
                        }, Icon('refresh')),
                        m('button.btn.btn-icon.btn-ghost', {
                            style: { color: 'var(--c-sidebar-text)', border: 'none', background: 'rgba(255,217,102,.1)' },
                            title: '退出登录',
                            onclick: () => {
                                authService.logout();
                                m.route.set('/login');
                            },
                        }, Icon('logout')),
                    ]),
                ]),
            ]),
        ];
    },
};
