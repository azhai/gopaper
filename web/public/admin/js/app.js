(function () {
  var PAGES = ['login', 'dashboard', 'content', 'analytics', 'editor', 'homepage', 'media', 'themes', 'settings'];
  var titles = {
    login: '登录',
    dashboard: '工作台',
    content: '内容',
    analytics: '统计',
    editor: '内容',
    homepage: '首页装修',
    media: '媒体库',
    themes: '主题设计',
    settings: '站点设置',
  };
  var templates = {};
  var mountedRoot = null;
  var booted = false;

  function queryString(query) {
    var parts = [];
    Object.keys(query || {}).forEach(function (key) {
      var value = query[key];
      if (value === undefined || value === null || value === '') return;
      parts.push(encodeURIComponent(key) + '=' + encodeURIComponent(value));
    });
    return parts.length ? '?' + parts.join('&') : '';
  }

  function joinPath(path, query) {
    return path + queryString(query);
  }

  function splitPath(value) {
    var index = value.indexOf('?');
    var path = index >= 0 ? value.slice(0, index) : value;
    var query = {};
    if (index >= 0) {
      new URLSearchParams(value.slice(index + 1)).forEach(function (v, k) { query[k] = v; });
    }
    return { path: path, query: query };
  }

  function parseLocation() {
    var path = location.pathname.replace(/^\/admin/, '') || '/';
    if (path.length > 1 && path.slice(-1) === '/') path = path.slice(0, -1);
    var query = {};
    new URLSearchParams(location.search).forEach(function (v, k) { query[k] = v; });
    return { path: path, query: query };
  }

  function route(name, path, query, params) {
    return { name: name, path: path, query: query || {}, params: params || {}, title: titles[name] || '管理后台' };
  }

  function resolve(path, query) {
    var legacy = {
      '/articles': ['/content', { type: 'article' }],
      '/pages': ['/content', { type: 'page' }],
      '/articles/new': ['/content/new', { type: 'article' }],
      '/pages/new': ['/content/new', { type: 'page' }],
      '/images': ['/media', {}],
      '/layout': ['/homepage', {}],
    };
    if (legacy[path]) return { redirect: legacy[path][0], query: legacy[path][1] };

    var match = path.match(/^\/articles\/([^/]+)\/edit$/);
    if (match) return { redirect: '/content/' + match[1] + '/edit', query: {} };
    match = path.match(/^\/pages\/([^/]+)\/edit$/);
    if (match) return { redirect: '/content/' + match[1] + '/edit', query: {} };

    if (path === '/' || path === '') path = '/dashboard';
    if (path === '/login') return route('login', path, query);
    if (path === '/dashboard') return route('dashboard', path, query);
    if (path === '/content') return route('content', path, query);
    if (path === '/analytics') return route('analytics', path, query);
    if (path === '/content/new') return route('editor', path, query);
    match = path.match(/^\/content\/([^/]+)\/edit$/);
    if (match) return route('editor', path, query, { slug: decodeURIComponent(match[1]) });
    if (path === '/homepage') return route('homepage', path, query);
    if (path === '/media') return route('media', path, query);
    if (path === '/themes') return route('themes', path, query);
    if (path === '/settings') return route('settings', path, query);
    return route('dashboard', '/dashboard', {});
  }

  function mount(name) {
    var rootId = name === 'login' ? 'login-root' : 'page-root';
    var root = document.getElementById(rootId);
    if (mountedRoot) {
      Alpine.destroyTree(mountedRoot);
      if (mountedRoot !== root) mountedRoot.innerHTML = '';
    }
    root.innerHTML = templates[name] || '';
    Alpine.initTree(root);
    mountedRoot = root;
  }

  function applyResolved(resolved, opts) {
    var logged = adminAuth.isLoggedIn();
    if (!logged && resolved.name !== 'login') {
      navigate('/login', { replace: true, skipGuard: true });
      return;
    }
    if (logged && resolved.name === 'login') {
      navigate('/dashboard', { replace: true, skipGuard: true });
      return;
    }
    Alpine.store('app').setRoute(resolved);
    mount(resolved.name);
    var url = '/admin' + resolved.path + queryString(resolved.query);
    if (opts.replace) history.replaceState({}, '', url);
    else history.pushState({}, '', url);
    document.title = resolved.title + ' - GoPaper 管理后台';
  }

  function navigate(path, opts) {
    opts = opts || {};
    var parsed = splitPath(path);
    var resolved = resolve(parsed.path, parsed.query);
    if (resolved.redirect) {
      navigate(joinPath(resolved.redirect, resolved.query), { replace: opts.replace, skipGuard: true });
      return;
    }

    var app = Alpine.store('app');
    if (!opts.skipGuard && app.guard && app.guard.fn()) {
      Alpine.store('overlay').confirm('离开当前页面？', app.guard.message, function () {
        app.clearGuard();
        navigate(path, { replace: opts.replace, skipGuard: true });
      });
      return;
    }

    if (mountedRoot && app.route.path === resolved.path && queryString(app.route.query) === queryString(resolved.query)) {
      app.closeSidebar();
      return;
    }

    applyResolved(resolved, { replace: !!opts.replace });
    app.closeSidebar();
    window.scrollTo(0, 0);
  }

  function onDocumentClick(e) {
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    var link = e.target.closest ? e.target.closest('a[href]') : null;
    if (!link || (link.target && link.target !== '_self')) return;
    var url;
    try { url = new URL(link.href, location.href); } catch (err) { return; }
    if (url.origin !== location.origin) return;
    if (url.pathname !== '/admin' && url.pathname.indexOf('/admin/') !== 0) return;
    e.preventDefault();
    navigate(url.pathname.slice('/admin'.length) + url.search);
  }

  function onPopState() {
    if (!booted) return;
    var parsed = parseLocation();
    var token = adminAuth.getToken();
    var app = Alpine.store('app');
    if (!token) {
      applyResolved(route('login', '/login', {}), { replace: true });
      return;
    }
    if (app.guard && app.guard.fn()) {
      var current = '/admin' + app.route.path + queryString(app.route.query);
      history.pushState({}, '', current);
      Alpine.store('overlay').confirm('离开当前页面？', app.guard.message, function () {
        app.clearGuard();
        var resolved = resolve(parsed.path, parsed.query);
        if (resolved.redirect) navigate(joinPath(resolved.redirect, resolved.query), { replace: true, skipGuard: true });
        else applyResolved(resolved, { replace: true });
      });
      return;
    }
    var resolved = resolve(parsed.path, parsed.query);
    if (resolved.redirect) navigate(joinPath(resolved.redirect, resolved.query), { replace: true, skipGuard: true });
    else applyResolved(resolved, { replace: true });
  }

  function preload() {
    return Promise.all(PAGES.map(function (name) {
      return fetch('/admin/pages/' + name + '.html', { cache: 'no-cache' })
        .then(function (response) {
          if (!response.ok) throw new Error(name);
          return response.text();
        })
        .then(function (html) { templates[name] = renderIcons(html); });
    }));
  }

  function boot() {
    preload().then(function () {
      booted = true;
      document.addEventListener('click', onDocumentClick, true);
      window.addEventListener('popstate', onPopState);
      var parsed = parseLocation();
      var resolved = resolve(parsed.path, parsed.query);
      if (resolved.redirect) navigate(joinPath(resolved.redirect, resolved.query), { replace: true, skipGuard: true });
      else applyResolved(resolved, { replace: true });
    }).catch(function (err) {
      var root = document.getElementById('page-root');
      if (root) {
        root.innerHTML = '<div class="error-card"><h3>后台加载失败</h3><p>页面资源加载失败，请刷新重试。</p></div>';
      }
      if (window.console) console.error('admin boot failed', err);
    });
  }

  window.adminNavigate = navigate;
  window.adminForceLogin = function () {
    var app = window.Alpine ? Alpine.store('app') : null;
    if (app) {
      app.loggedIn = false;
      app.clearGuard();
    }
    if (booted) navigate('/login', { replace: true, skipGuard: true });
    else location.href = '/admin/login';
  };

  document.addEventListener('alpine:initialized', boot);
})();
