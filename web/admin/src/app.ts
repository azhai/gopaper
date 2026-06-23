import './styles/admin.css';
import m from 'mithril';
import { authService } from './services/apiService';
import { LoginPage } from './routes/login';
import { DashboardPage } from './routes/dashboard';
import { ArticlesPage } from './routes/articles';
import { PagesPage } from './routes/pages';
import { EditorPage } from './routes/editor';
import { ImagesPage } from './routes/images';
import { SettingsPage } from './routes/settings';
import { LayoutPage } from './routes/layout';
import { Sidebar, topbarState } from './components/navbar';
import { Overlay } from './components/overlay';
import { Toast } from './components/toast';
import { Icon } from './icons';

const titles: Record<string, string> = {
  '/dashboard': '仪表盘',
  '/articles': '文章管理',
  '/pages': '页面管理',
  '/articles/new': '新建文章',
  '/pages/new': '新建页面',
  '/images': '图片管理',
  '/settings': '站点设置',
  '/layout': '布局管理',
};

const Layout: m.Component = {
  view(vnode) {
    if (!authService.isLoggedIn()) return vnode.children;
    const current = m.route.get();
    const title = topbarState.title || titles[current] || '管理后台';
    const crumb = topbarState.crumb;
    return m('.app-shell', [
      m(Sidebar),
      m('.app-main', [
        m('header.topbar', [
          m('.topbar-title', [
            m('h1', title),
            crumb ? m('.crumb', crumb) : null,
          ]),
          m('.topbar-actions', topbarState.actions || [
            m('a.btn.btn-ghost.btn-sm', { href: '/', target: '_blank' }, [
              Icon('external'), '访问站点',
            ]),
          ]),
        ]),
        m('main.content', vnode.children),
      ]),
      m(Overlay),
      m(Toast),
    ]);
  },
};

const guard = (component: m.Component): m.RouteResolver => ({
  onmatch() {
    if (!authService.isLoggedIn()) {
      m.route.set('/login');
      return;
    }
    return component;
  },
  render(vnode: m.Vnode) {
    return m(Layout, vnode);
  },
});

const guardEditor = (mode: 'page' | 'article'): m.RouteResolver => ({
  onmatch() {
    if (!authService.isLoggedIn()) {
      m.route.set('/login');
      return;
    }
    return EditorPage;
  },
  render(vnode: m.Vnode) {
    return m(Layout, m(EditorPage, { ...vnode.attrs, mode }));
  },
});

m.route(document.body!, '/dashboard', {
  '/login': {
    render() {
      return m(LoginPage);
    },
  },
  '/dashboard': guard(DashboardPage),
  '/articles': guard(ArticlesPage),
  '/articles/new': guardEditor('article'),
  '/articles/:slug/edit': guardEditor('article'),
  '/pages': guard(PagesPage),
  '/pages/new': guardEditor('page'),
  '/pages/:slug/edit': guardEditor('page'),
  '/images': guard(ImagesPage),
  '/settings': guard(SettingsPage),
  '/layout': guard(LayoutPage),
});
